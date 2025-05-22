"""
Worker module for AI task execution in the AI-Job scheduling system.
This worker connects to the Go scheduler and executes LLM tasks.
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
from transformers import AutoModelForCausalLM, AutoTokenizer, pipeline

# Configure logging
logger = logging.getLogger("ai-worker")
logger.setLevel(logging.INFO)

# Create formatter
formatter = logging.Formatter(
    '{"time": "%(asctime)s", "name": "%(name)s", "level": "%(levelname)s", "message": "%(message)s"}',
    datefmt="%Y-%m-%dT%H:%M:%S%z"
)

# Console handler
console_handler = logging.StreamHandler(sys.stdout)
console_handler.setFormatter(formatter)

# File handler
log_dir = "/var/log/app"
os.makedirs(log_dir, exist_ok=True)
file_handler = logging.FileHandler(f"{log_dir}/ai-worker.log")
file_handler.setFormatter(formatter)

# Add handlers
logger.addHandler(console_handler)
logger.addHandler(file_handler)

class Config:
    """Configuration for the worker."""
    
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
        
        # LLM config
        self.models = config.get("llm", {}).get("models", [])

class Worker:
    """Worker for executing AI tasks."""
    
    def __init__(self, config: Config):
        """Initialize the worker."""
        self.config = config
        self.worker_id = None
        self.name = f"worker-{platform.node()}-{uuid.uuid4().hex[:8]}"
        self.capabilities = ["text-generation", "embeddings"]
        self.current_task_id = None
        self.models = {}
        self.running = True
        
        # Setup signal handlers
        signal.signal(signal.SIGINT, self._handle_signal)
        signal.signal(signal.SIGTERM, self._handle_signal)
        
        logger.info(f"Initializing worker {self.name}")
    
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
            logger.info(f"Worker registered with ID: {self.worker_id}")
            return True
        except Exception as e:
            logger.error(f"Failed to register worker: {e}")
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
    
    def update_status(self, status: str, task_status: Optional[str] = None, 
                    task_output: Optional[bytes] = None, task_error: Optional[str] = None):
        """Update the worker status on the server."""
        if not self.worker_id:
            return False
        
        available_memory, available_cpu, available_gpu = self._get_resource_metrics()
        
        data = {
            "status": status,
            "current_task_id": self.current_task_id,
            "available_memory": available_memory,
            "available_cpu": available_cpu,
            "available_gpu": available_gpu
        }
        
        if task_status:
            data["task_status"] = task_status
        
        if task_output:
            data["task_output"] = task_output.decode("utf-8") if isinstance(task_output, bytes) else task_output
        
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
    
    def get_task(self, task_id: str):
        """Get a task by ID."""
        try:
            response = requests.get(
                f"{self.config.api_base_url}/api/v1/tasks/{task_id}"
            )
            response.raise_for_status()
            return response.json()
        except Exception as e:
            logger.error(f"Failed to get task {task_id}: {e}")
            return None
    
    def _get_resource_metrics(self):
        """Get resource metrics for the worker."""
        # Memory
        memory = psutil.virtual_memory()
        available_memory = memory.available
        
        # CPU
        available_cpu = psutil.cpu_percent(interval=0.1) / 100.0
        available_cpu = 1.0 - available_cpu  # Convert to available percentage
        
        # GPU
        available_gpu = 0.0
        if torch.cuda.is_available():
            # This is a simplistic way to report GPU availability
            # Real implementation should use nvidia-smi or similar
            try:
                gpu_memory = torch.cuda.get_device_properties(0).total_memory
                allocated_memory = torch.cuda.memory_allocated(0)
                available_gpu = 1.0 - (allocated_memory / gpu_memory)
            except Exception as e:
                logger.warning(f"Failed to get GPU metrics: {e}")
        
        return available_memory, available_cpu, available_gpu
    
    def load_model(self, model_name: str):
        """Load a model by name."""
        if model_name in self.models:
            return self.models[model_name]
        
        # Find model config
        model_config = None
        for model in self.config.models:
            if model["name"] == model_name:
                model_config = model
                break
        
        if not model_config:
            raise ValueError(f"Model {model_name} not found in configuration")
        
        logger.info(f"Loading model {model_name}")
        
        # Load the model based on provider
        if model_config["provider"] == "local":
            # Local model loading
            model_path = model_config["model_path"]
            
            # Check for quantization
            kwargs = {}
            if "quantization" in model_config:
                if model_config["quantization"] == "int8":
                    kwargs["load_in_8bit"] = True
                elif model_config["quantization"] == "int4":
                    kwargs["load_in_4bit"] = True
            
            device = "cuda" if torch.cuda.is_available() else "cpu"
            logger.info(f"Loading model from {model_path} on {device}")
            
            model = AutoModelForCausalLM.from_pretrained(
                model_path,
                device_map=device,
                **kwargs
            )
            tokenizer = AutoTokenizer.from_pretrained(model_path)
            
            self.models[model_name] = {
                "model": model,
                "tokenizer": tokenizer,
                "pipeline": pipeline(
                    "text-generation",
                    model=model,
                    tokenizer=tokenizer,
                    device=0 if device == "cuda" else -1
                )
            }
            
            logger.info(f"Model {model_name} loaded successfully")
            return self.models[model_name]
        
        elif model_config["provider"] == "openai":
            # For OpenAI models, we just store the config - we'll use the API
            self.models[model_name] = {
                "config": model_config,
                "provider": "openai"
            }
            logger.info(f"OpenAI model {model_name} registered")
            return self.models[model_name]
        
        else:
            raise ValueError(f"Unsupported model provider: {model_config['provider']}")
    
    def execute_task(self, task: Dict[str, Any]):
        """Execute a task."""
        task_id = task["id"]
        model_name = task["model_name"]
        
        logger.info(f"Executing task {task_id} with model {model_name}")
        
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
            
            # Load the model if not already loaded
            model_data = self.load_model(model_name)
            
            # Execute based on model provider
            if model_data.get("provider") == "openai":
                output = self._run_openai_task(model_data, task_input)
            else:
                output = self._run_local_model_task(model_data, task_input)
            
            # Convert output to JSON string and then to bytes
            output_bytes = json.dumps(output).encode("utf-8")
            
            # Update task status to completed
            self.update_status(
                "available", 
                task_status="completed", 
                task_output=output_bytes
            )
            
            logger.info(f"Task {task_id} completed successfully")
            
        except Exception as e:
            logger.error(f"Error executing task {task_id}: {e}", exc_info=True)
            self.update_status(
                "available", 
                task_status="failed", 
                task_error=str(e)
            )
        
        finally:
            self.current_task_id = None
    
    def _run_local_model_task(self, model_data, task_input):
        """Run a task with a local model."""
        prompt = task_input.get("prompt", "")
        max_length = task_input.get("max_tokens", 1024)
        temperature = task_input.get("temperature", 0.7)
        
        # Run the model inference
        result = model_data["pipeline"](
            prompt,
            max_length=max_length,
            temperature=temperature,
            do_sample=temperature > 0,
            num_return_sequences=1
        )
        
        # Extract and return the generated text
        generated_text = result[0].get("generated_text", "")
        if prompt in generated_text:
            generated_text = generated_text[len(prompt):]
        
        return {"text": generated_text}
    
    def _run_openai_task(self, model_data, task_input):
        """Run a task with an OpenAI model - this is a placeholder that would use an API."""
        # In a real implementation, this would use the OpenAI API
        logger.info("OpenAI API integration would go here")
        
        # Return a placeholder response
        return {"text": "This is a placeholder for OpenAI API integration"}
    
    def run(self):
        """Run the worker main loop."""
        logger.info("Starting worker main loop")
        
        last_heartbeat = 0
        
        while self.running:
            current_time = time.time()
            
            # Send heartbeat at regular intervals
            if current_time - last_heartbeat > self.config.heartbeat_interval:
                self.heartbeat()
                last_heartbeat = current_time
            
            # If we're not busy, we're available for tasks
            if not self.current_task_id:
                self.update_status("available")
                
                # Check for assigned tasks (this would be via polling in this simple implementation)
                # In a more advanced implementation, this could use webhooks or a message queue
                try:
                    response = requests.get(
                        f"{self.config.api_base_url}/api/v1/tasks",
                        params={"status": "scheduled"}
                    )
                    response.raise_for_status()
                    
                    tasks = response.json()
                    for task in tasks:
                        if task.get("worker_id") == self.worker_id:
                            logger.info(f"Found assigned task: {task['id']}")
                            self.execute_task(task)
                            break
                
                except Exception as e:
                    logger.error(f"Error polling for tasks: {e}")
            
            # Sleep to avoid hammering the API
            time.sleep(self.config.poll_interval)

def main():
    """Main entry point for the worker."""
    parser = argparse.ArgumentParser(description="AI task worker")
    parser.add_argument(
        "--config", 
        default="../config/config.yaml", 
        help="Path to configuration file"
    )
    args = parser.parse_args()
    
    try:
        config = Config(args.config)
        worker = Worker(config)
        
        if worker.register():
            worker.run()
        else:
            logger.error("Failed to register worker, exiting")
            sys.exit(1)
    
    except Exception as e:
        logger.error(f"Unhandled exception: {e}", exc_info=True)
        sys.exit(1)

if __name__ == "__main__":
    main()