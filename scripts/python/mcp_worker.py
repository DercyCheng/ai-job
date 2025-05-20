"""
MCP Worker module for executing MCP-specific tasks in the AI-Job scheduling system.
This worker connects to the MCP server and processes MCP tasks.
"""

import argparse
import json
import logging
import os
import platform
import signal
import sys
import time
import uuid
from datetime import datetime
from typing import Dict, List, Optional, Union, Any

import psutil
import requests
import torch
import yaml

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger("mcp-worker")

class Config:
    """Configuration for the MCP worker."""
    
    def __init__(self, config_path: str):
        """Initialize the configuration."""
        with open(config_path, "r") as f:
            config = yaml.safe_load(f)
        
        # Server config
        server_config = config.get("server", {})
        self.api_base_url = f"http://{server_config.get('host', 'localhost')}:{server_config.get('port', 8080)}"
        
        # Worker config
        worker_config = config.get("worker", {})
        self.heartbeat_interval = worker_config.get("heartbeat_interval", 10)
        self.poll_interval = worker_config.get("poll_interval", 1)
        
        # MCP config
        mcp_config = config.get("mcp", {})
        self.mcp_server_url = mcp_config.get("server_url", "http://localhost:8001")
        self.mcp_enabled = mcp_config.get("enabled", False)

class MCPWorker:
    """Worker for executing MCP tasks."""
    
    def __init__(self, config: Config):
        """Initialize the worker."""
        self.config = config
        self.worker_id = None
        self.name = f"mcp-worker-{platform.node()}-{uuid.uuid4().hex[:8]}"
        self.capabilities = ["mcp-tasks"]
        self.current_task_id = None
        self.running = True
        
        # Setup signal handlers
        signal.signal(signal.SIGINT, self._handle_signal)
        signal.signal(signal.SIGTERM, self._handle_signal)
        
        logger.info(f"Initializing MCP worker {self.name}")
    
    def _handle_signal(self, sig, frame):
        """Handle termination signals."""
        logger.info(f"Received signal {sig}, shutting down...")
        self.running = False
    
    def register(self):
        """Register the worker with the server."""
        available_memory, available_cpu, available_gpu = self._get_resource_metrics()
        
        data = {
            "name": self.name,
            "capabilities": self.capabilities,
            "available_memory": available_memory,
            "available_cpu": available_cpu,
            "available_gpu": available_gpu
        }
        
        try:
            response = requests.post(
                f"{self.config.api_base_url}/api/v1/workers",
                json=data
            )
            response.raise_for_status()
            self.worker_id = response.json().get("id")
            logger.info(f"MCP Worker registered with ID: {self.worker_id}")
            return True
        except Exception as e:
            logger.error(f"Failed to register MCP worker: {e}")
            return False
    
    def heartbeat(self):
        """Send a heartbeat to the server."""
        if not self.worker_id:
            return False
        
        try:
            response = requests.put(
                f"{self.config.api_base_url}/api/v1/workers/{self.worker_id}/heartbeat"
            )
            response.raise_for_status()
            return True
        except Exception as e:
            logger.error(f"Failed to send heartbeat: {e}")
            return False
    
    def update_status(self, status: str, task_status: str = None, task_output: bytes = None, task_error: str = None):
        """Update worker status and optionally task status."""
        if not self.worker_id:
            return False
        
        available_memory, available_cpu, available_gpu = self._get_resource_metrics()
        
        data = {
            "status": status,
            "available_memory": available_memory,
            "available_cpu": available_cpu,
            "available_gpu": available_gpu
        }
        
        if self.current_task_id:
            data["current_task_id"] = self.current_task_id
            
            if task_status:
                data["task_status"] = task_status
            
            if task_output:
                data["task_output"] = task_output
            
            if task_error:
                data["task_error"] = task_error
        
        try:
            response = requests.put(
                f"{self.config.api_base_url}/api/v1/workers/{self.worker_id}/status",
                json=data
            )
            response.raise_for_status()
            return True
        except Exception as e:
            logger.error(f"Failed to update status: {e}")
            return False
    
    def _get_resource_metrics(self):
        """Get resource metrics for the worker."""
        # Get memory info
        memory = psutil.virtual_memory()
        available_memory = memory.available
        
        # Get CPU info
        available_cpu = psutil.cpu_percent(interval=0.1)
        
        # Get GPU info if available
        available_gpu = 0
        if torch.cuda.is_available():
            available_gpu = 100 - torch.cuda.utilization(0)
        
        return available_memory, available_cpu, available_gpu
    
    def execute_mcp_task(self, task: Dict[str, Any]):
        """Execute an MCP task."""
        task_id = task["id"]
        task_type = task["type"]
        
        logger.info(f"Executing MCP task {task_id} of type {task_type}")
        
        # Update task status to running
        self.current_task_id = task_id
        self.update_status("busy", task_status="running")
        
        try:
            # Parse the input
            task_input = task["input"]
            if isinstance(task_input, bytes):
                task_input = task_input.decode("utf-8")
            if isinstance(task_input, str):
                task_input = json.loads(task_input)
            
            # Process the task based on type
            if task_type == "mcp.create_context":
                output = self._handle_create_context(task_input)
            elif task_type == "mcp.add_prompt":
                output = self._handle_add_prompt(task_input)
            elif task_type == "mcp.add_node":
                output = self._handle_add_node(task_input)
            elif task_type == "mcp.delete_node":
                output = self._handle_delete_node(task_input)
            elif task_type == "mcp.delete_context":
                output = self._handle_delete_context(task_input)
            else:
                raise ValueError(f"Unsupported MCP task type: {task_type}")
            
            # Convert output to JSON string and then to bytes
            output_bytes = json.dumps(output).encode("utf-8")
            
            # Update task status to completed
            self.update_status(
                "available", 
                task_status="completed", 
                task_output=output_bytes
            )
            
            logger.info(f"MCP task {task_id} completed successfully")
            
        except Exception as e:
            logger.error(f"Error executing MCP task {task_id}: {e}", exc_info=True)
            self.update_status(
                "available", 
                task_status="failed", 
                task_error=str(e)
            )
        
        finally:
            self.current_task_id = None
    
    def _handle_create_context(self, task_input: Dict[str, Any]) -> Dict[str, Any]:
        """Handle creating a context in the MCP server."""
        logger.info("Creating MCP context")
        
        model_id = task_input.get("model_id")
        nodes = task_input.get("nodes", [])
        metadata = task_input.get("metadata", {})
        return_context = task_input.get("return_context", False)
        
        # Create context request
        request_data = {
            "model_id": model_id,
            "nodes": nodes,
            "metadata": metadata,
            "return_context": return_context
        }
        
        # Send request to MCP server
        response = requests.post(
            f"{self.config.mcp_server_url}/v1/contexts",
            json=request_data
        )
        response.raise_for_status()
        
        # Return the response data
        return response.json()
    
    def _handle_add_prompt(self, task_input: Dict[str, Any]) -> Dict[str, Any]:
        """Handle adding a prompt to a context in the MCP server."""
        logger.info("Adding prompt to MCP context")
        
        context_id = task_input.get("context_id")
        prompt = task_input.get("prompt")
        prompt_id = task_input.get("prompt_id")
        parent_id = task_input.get("parent_id")
        metadata = task_input.get("metadata", {})
        stream = task_input.get("stream", False)
        
        # Create prompt request
        request_data = {
            "prompt": prompt,
            "metadata": metadata,
            "stream": stream
        }
        
        if prompt_id:
            request_data["prompt_id"] = prompt_id
            
        if parent_id:
            request_data["parent_id"] = parent_id
        
        # Send request to MCP server
        response = requests.post(
            f"{self.config.mcp_server_url}/v1/contexts/{context_id}/prompt",
            json=request_data
        )
        response.raise_for_status()
        
        # Return the response data
        return response.json()
    
    def _handle_add_node(self, task_input: Dict[str, Any]) -> Dict[str, Any]:
        """Handle adding a node to a context in the MCP server."""
        logger.info("Adding node to MCP context")
        
        context_id = task_input.get("context_id")
        node = task_input.get("node", {})
        
        # Send request to MCP server
        response = requests.post(
            f"{self.config.mcp_server_url}/v1/contexts/{context_id}/nodes",
            json=node
        )
        response.raise_for_status()
        
        # Return the response data
        return response.json()
    
    def _handle_delete_node(self, task_input: Dict[str, Any]) -> Dict[str, Any]:
        """Handle deleting a node from a context in the MCP server."""
        logger.info("Deleting node from MCP context")
        
        context_id = task_input.get("context_id")
        node_id = task_input.get("node_id")
        
        # Send request to MCP server
        response = requests.delete(
            f"{self.config.mcp_server_url}/v1/contexts/{context_id}/nodes/{node_id}"
        )
        response.raise_for_status()
        
        # Return the response data
        return response.json()
    
    def _handle_delete_context(self, task_input: Dict[str, Any]) -> Dict[str, Any]:
        """Handle deleting a context in the MCP server."""
        logger.info("Deleting MCP context")
        
        context_id = task_input.get("context_id")
        
        # Send request to MCP server
        response = requests.delete(
            f"{self.config.mcp_server_url}/v1/contexts/{context_id}"
        )
        response.raise_for_status()
        
        # Return the response data
        return response.json()
    
    def run(self):
        """Run the worker."""
        logger.info("MCP Worker starting")
        
        last_heartbeat = 0
        
        while self.running:
            # Send heartbeat if needed
            current_time = time.time()
            if current_time - last_heartbeat > self.config.heartbeat_interval:
                if self.heartbeat():
                    last_heartbeat = current_time
            
            # Check for tasks if not currently processing one
            if not self.current_task_id:
                self.update_status("available")
                
                # Check for MCP tasks
                try:
                    response = requests.get(
                        f"{self.config.api_base_url}/api/v1/mcp/tasks",
                        params={"status": "scheduled"}
                    )
                    response.raise_for_status()
                    
                    tasks = response.json()
                    for task in tasks:
                        if task.get("worker_id") == self.worker_id:
                            logger.info(f"Found assigned MCP task: {task['id']}")
                            self.execute_mcp_task(task)
                            break
                
                except Exception as e:
                    logger.error(f"Error polling for MCP tasks: {e}")
            
            # Sleep to avoid hammering the API
            time.sleep(self.config.poll_interval)

def main():
    """Main entry point for the MCP worker."""
    parser = argparse.ArgumentParser(description="MCP task worker")
    parser.add_argument(
        "--config", 
        default="../config/config.yaml", 
        help="Path to configuration file"
    )
    args = parser.parse_args()
    
    try:
        config = Config(args.config)
        
        if not config.mcp_enabled:
            logger.info("MCP is not enabled in the configuration, exiting")
            sys.exit(0)
            
        worker = MCPWorker(config)
        
        if worker.register():
            worker.run()
        else:
            logger.error("Failed to register MCP worker, exiting")
            sys.exit(1)
    
    except Exception as e:
        logger.error(f"Unhandled exception: {e}", exc_info=True)
        sys.exit(1)

if __name__ == "__main__":
    main()
