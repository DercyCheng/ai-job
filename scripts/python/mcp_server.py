"""
Model Context Protocol Server for AI-Job system.
Implements the MCP specification for standardized model interaction.
"""

import argparse
import json
import logging
import os
import sys
import time
import uuid
from typing import Dict, List, Optional, Union, Any, Tuple

import torch
import uvicorn
from fastapi import FastAPI, HTTPException, Request, Response, Body, Depends
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from transformers import AutoModelForCausalLM, AutoTokenizer, pipeline

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger("mcp-server")

# Create FastAPI app
app = FastAPI(title="Model Context Protocol Server")

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Global variables for loaded models and contexts
loaded_models = {}
active_contexts = {}
active_contexts_lock = {}  # Used to prevent concurrent modifications to the same context

# MCP Protocol Models (based on MCP specification)

class MCPContextNode(BaseModel):
    """Node in the MCP context tree."""
    id: str
    content: str
    content_type: str = "text/plain"
    metadata: Dict[str, Any] = Field(default_factory=dict)
    parent: Optional[str] = None
    children: List[str] = Field(default_factory=list)

class MCPModel(BaseModel):
    """Model information in MCP."""
    id: str
    name: str
    provider: str
    capabilities: List[str] = Field(default_factory=list)
    config: Dict[str, Any] = Field(default_factory=dict)

class MCPContextCreateRequest(BaseModel):
    """Request to create a new context."""
    model_id: str
    nodes: List[MCPContextNode] = Field(default_factory=list)
    metadata: Dict[str, Any] = Field(default_factory=dict)
    return_context: bool = True

class MCPContextCreateResponse(BaseModel):
    """Response after creating a context."""
    context_id: str
    model: MCPModel
    nodes: List[MCPContextNode] = Field(default_factory=list)
    metadata: Dict[str, Any] = Field(default_factory=dict)

class MCPPromptRequest(BaseModel):
    """Request to add a prompt to a context."""
    context_id: str
    prompt: str
    prompt_id: Optional[str] = None
    parent_id: Optional[str] = None
    metadata: Dict[str, Any] = Field(default_factory=dict)
    stream: bool = False

class MCPPromptResponse(BaseModel):
    """Response to a prompt request."""
    context_id: str
    prompt_id: str
    completion_id: str
    completion: str
    metadata: Dict[str, Any] = Field(default_factory=dict)

class MCPStreamChunk(BaseModel):
    """Chunk of a streaming response."""
    context_id: str
    prompt_id: str
    completion_id: str
    completion_chunk: str
    is_final: bool = False
    metadata: Dict[str, Any] = Field(default_factory=dict)

class MCPAddNodeRequest(BaseModel):
    """Request to add a node to a context."""
    context_id: str
    node: MCPContextNode

class MCPAddNodeResponse(BaseModel):
    """Response after adding a node."""
    context_id: str
    node: MCPContextNode

class MCPDeleteNodeRequest(BaseModel):
    """Request to delete a node from a context."""
    context_id: str
    node_id: str

class MCPDeleteNodeResponse(BaseModel):
    """Response after deleting a node."""
    context_id: str
    deleted: bool

class MCPListContextsResponse(BaseModel):
    """Response with list of contexts."""
    contexts: List[str]

class MCPGetContextResponse(BaseModel):
    """Response with context details."""
    context_id: str
    model: MCPModel
    nodes: List[MCPContextNode]
    metadata: Dict[str, Any]

class MCPDeleteContextResponse(BaseModel):
    """Response after deleting a context."""
    context_id: str
    deleted: bool

# Model loading function, similar to the llm_server implementation
def load_model(model_name: str) -> Dict:
    """Load a model and tokenizer."""
    if model_name in loaded_models:
        logger.info(f"Using already loaded model: {model_name}")
        return loaded_models[model_name]
    
    logger.info(f"Loading model: {model_name}")
    
    # Check if model exists in the local filesystem or is a huggingface model
    model_path = model_name
    if not os.path.exists(model_name):
        # Assume it's a HuggingFace model
        model_path = model_name
    
    # Determine device
    device = "cuda" if torch.cuda.is_available() else "cpu"
    logger.info(f"Using device: {device}")
    
    # Load the model with optimizations appropriate for the model size
    kwargs = {}
    if "int8" in model_name.lower():
        kwargs["load_in_8bit"] = True
    elif "int4" in model_name.lower():
        kwargs["load_in_4bit"] = True
    
    start_time = time.time()
    
    try:
        model = AutoModelForCausalLM.from_pretrained(
            model_path,
            device_map=device,
            **kwargs
        )
        tokenizer = AutoTokenizer.from_pretrained(model_path)
        
        # Create the generation pipeline
        gen_pipeline = pipeline(
            "text-generation",
            model=model,
            tokenizer=tokenizer,
            device=0 if device == "cuda" else -1
        )
        
        logger.info(f"Model loaded in {time.time() - start_time:.2f} seconds")
        
        # Store the model data
        model_data = {
            "model": model,
            "tokenizer": tokenizer,
            "pipeline": gen_pipeline,
            "device": device,
            "loaded_at": time.time(),
            "config": {
                "max_length": model.config.max_position_embeddings if hasattr(model.config, "max_position_embeddings") else 2048,
                "model_type": model.config.model_type if hasattr(model.config, "model_type") else "unknown",
            }
        }
        
        loaded_models[model_name] = model_data
        return model_data
    
    except Exception as e:
        logger.error(f"Error loading model {model_name}: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to load model: {str(e)}")

def get_node_by_id(context_id: str, node_id: str) -> Optional[MCPContextNode]:
    """Get a node from a context by ID."""
    if context_id not in active_contexts:
        return None
    
    context = active_contexts[context_id]
    for node in context["nodes"]:
        if node.id == node_id:
            return node
    
    return None

def build_context_tree(context_id: str) -> List[Tuple[str, str]]:
    """Build a flattened representation of the context tree for generation."""
    if context_id not in active_contexts:
        return []
    
    context = active_contexts[context_id]
    nodes = {node.id: node for node in context["nodes"]}
    
    # Find root nodes (nodes without parents)
    root_nodes = [node for node in context["nodes"] if node.parent is None]
    
    # Build the flattened tree using DFS
    flattened_tree = []
    
    def dfs(node):
        flattened_tree.append((node.id, node.content))
        for child_id in node.children:
            if child_id in nodes:
                dfs(nodes[child_id])
    
    for root in root_nodes:
        dfs(root)
    
    return flattened_tree

def build_prompt_from_context(context_id: str) -> str:
    """Build a prompt string from the context tree."""
    tree = build_context_tree(context_id)
    return "\n".join([content for _, content in tree])

# MCP API Endpoints

@app.post("/v1/contexts", response_model=MCPContextCreateResponse)
async def create_context(request: MCPContextCreateRequest):
    """Create a new context with the specified model."""
    try:
        # Load the model
        model_data = load_model(request.model_id)
        
        # Generate a context ID
        context_id = str(uuid.uuid4())
        
        # Create the context
        context = {
            "model_id": request.model_id,
            "model_data": model_data,
            "nodes": request.nodes,
            "metadata": request.metadata,
            "created_at": time.time(),
            "updated_at": time.time()
        }
        
        # Store the context
        active_contexts[context_id] = context
        active_contexts_lock[context_id] = False
        
        # Prepare the response
        model = MCPModel(
            id=request.model_id,
            name=request.model_id,
            provider="local",
            capabilities=["text-generation", "embeddings"],
            config=model_data["config"]
        )
        
        response = MCPContextCreateResponse(
            context_id=context_id,
            model=model,
            nodes=request.nodes if request.return_context else [],
            metadata=request.metadata
        )
        
        logger.info(f"Created context {context_id} with model {request.model_id}")
        return response
    
    except Exception as e:
        logger.error(f"Error creating context: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to create context: {str(e)}")

@app.post("/v1/contexts/{context_id}/prompt", response_model=MCPPromptResponse)
async def add_prompt(context_id: str, request: MCPPromptRequest):
    """Add a prompt to an existing context and generate a completion."""
    try:
        if context_id not in active_contexts:
            raise HTTPException(status_code=404, detail=f"Context {context_id} not found")
        
        # Check if the context is locked (another operation in progress)
        if active_contexts_lock.get(context_id, False):
            raise HTTPException(status_code=409, detail=f"Context {context_id} is currently being modified by another request")
        
        # Lock the context
        active_contexts_lock[context_id] = True
        
        try:
            context = active_contexts[context_id]
            model_data = context["model_data"]
            
            # Generate IDs if not provided
            prompt_id = request.prompt_id or str(uuid.uuid4())
            completion_id = str(uuid.uuid4())
            
            # Create a prompt node
            prompt_node = MCPContextNode(
                id=prompt_id,
                content=request.prompt,
                content_type="text/plain",
                metadata=request.metadata,
                parent=request.parent_id,
                children=[completion_id]
            )
            
            # Add the prompt node to the context
            context["nodes"].append(prompt_node)
            
            # If parent ID is specified, update the parent's children
            if request.parent_id:
                parent_node = get_node_by_id(context_id, request.parent_id)
                if parent_node:
                    parent_node.children.append(prompt_id)
            
            # Build the context for generation
            full_prompt = build_prompt_from_context(context_id)
            
            # Generate a completion
            generation_kwargs = {
                "max_length": len(model_data["tokenizer"].encode(full_prompt)) + 512,  # Default to 512 tokens
                "temperature": 0.7,
                "do_sample": True,
                "top_p": 0.9,
                "top_k": 50,
                "num_return_sequences": 1
            }
            
            # Update kwargs from metadata if provided
            if "generation_params" in request.metadata:
                generation_kwargs.update(request.metadata["generation_params"])
            
            # Handle streaming if requested
            if request.stream:
                # This would normally set up streaming, but for now we'll generate everything
                # To properly implement streaming, you would need to use the tokenizer to generate
                # token by token and yield each token as a separate response
                pass
            
            # Generate text
            result = model_data["pipeline"](
                full_prompt,
                **generation_kwargs
            )
            
            # Extract generated text
            generated_text = result[0]["generated_text"]
            
            # Remove the prompt from the beginning of the generated text
            if full_prompt in generated_text:
                generated_text = generated_text[len(full_prompt):]
            
            # Create a completion node
            completion_node = MCPContextNode(
                id=completion_id,
                content=generated_text,
                content_type="text/plain",
                metadata={
                    "generation_time": time.time(),
                    "model_id": context["model_id"],
                },
                parent=prompt_id,
                children=[]
            )
            
            # Add the completion node to the context
            context["nodes"].append(completion_node)
            
            # Update the context's updated_at timestamp
            context["updated_at"] = time.time()
            
            # Prepare the response
            response = MCPPromptResponse(
                context_id=context_id,
                prompt_id=prompt_id,
                completion_id=completion_id,
                completion=generated_text,
                metadata={
                    "tokens": len(model_data["tokenizer"].encode(generated_text)),
                    "generation_time": time.time()
                }
            )
            
            logger.info(f"Generated completion for prompt {prompt_id} in context {context_id}")
            return response
        
        finally:
            # Unlock the context
            active_contexts_lock[context_id] = False
    
    except HTTPException:
        # Re-raise HTTP exceptions
        raise
    
    except Exception as e:
        logger.error(f"Error processing prompt: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to process prompt: {str(e)}")

@app.post("/v1/contexts/{context_id}/nodes", response_model=MCPAddNodeResponse)
async def add_node(context_id: str, request: MCPAddNodeRequest):
    """Add a node to an existing context."""
    try:
        if context_id not in active_contexts:
            raise HTTPException(status_code=404, detail=f"Context {context_id} not found")
        
        if active_contexts_lock.get(context_id, False):
            raise HTTPException(status_code=409, detail=f"Context {context_id} is currently being modified by another request")
        
        active_contexts_lock[context_id] = True
        
        try:
            context = active_contexts[context_id]
            
            # Add the node to the context
            context["nodes"].append(request.node)
            
            # If parent ID is specified, update the parent's children
            if request.node.parent:
                parent_node = get_node_by_id(context_id, request.node.parent)
                if parent_node:
                    parent_node.children.append(request.node.id)
            
            # Update the context's updated_at timestamp
            context["updated_at"] = time.time()
            
            # Return the added node
            return MCPAddNodeResponse(context_id=context_id, node=request.node)
        
        finally:
            active_contexts_lock[context_id] = False
    
    except HTTPException:
        raise
    
    except Exception as e:
        logger.error(f"Error adding node: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to add node: {str(e)}")

@app.delete("/v1/contexts/{context_id}/nodes/{node_id}", response_model=MCPDeleteNodeResponse)
async def delete_node(context_id: str, node_id: str):
    """Delete a node from an existing context."""
    try:
        if context_id not in active_contexts:
            raise HTTPException(status_code=404, detail=f"Context {context_id} not found")
        
        if active_contexts_lock.get(context_id, False):
            raise HTTPException(status_code=409, detail=f"Context {context_id} is currently being modified by another request")
        
        active_contexts_lock[context_id] = True
        
        try:
            context = active_contexts[context_id]
            
            # Find the node
            node_index = None
            for i, node in enumerate(context["nodes"]):
                if node.id == node_id:
                    node_index = i
                    break
            
            if node_index is None:
                return MCPDeleteNodeResponse(context_id=context_id, deleted=False)
            
            node = context["nodes"][node_index]
            
            # Remove the node from its parent's children
            if node.parent:
                parent_node = get_node_by_id(context_id, node.parent)
                if parent_node and node_id in parent_node.children:
                    parent_node.children.remove(node_id)
            
            # Remove the node from the context
            del context["nodes"][node_index]
            
            # Update the context's updated_at timestamp
            context["updated_at"] = time.time()
            
            return MCPDeleteNodeResponse(context_id=context_id, deleted=True)
        
        finally:
            active_contexts_lock[context_id] = False
    
    except HTTPException:
        raise
    
    except Exception as e:
        logger.error(f"Error deleting node: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to delete node: {str(e)}")

@app.get("/v1/contexts", response_model=MCPListContextsResponse)
async def list_contexts():
    """List all active contexts."""
    try:
        return MCPListContextsResponse(contexts=list(active_contexts.keys()))
    
    except Exception as e:
        logger.error(f"Error listing contexts: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to list contexts: {str(e)}")

@app.get("/v1/contexts/{context_id}", response_model=MCPGetContextResponse)
async def get_context(context_id: str):
    """Get a context by ID."""
    try:
        if context_id not in active_contexts:
            raise HTTPException(status_code=404, detail=f"Context {context_id} not found")
        
        context = active_contexts[context_id]
        
        model = MCPModel(
            id=context["model_id"],
            name=context["model_id"],
            provider="local",
            capabilities=["text-generation", "embeddings"],
            config=context["model_data"]["config"]
        )
        
        return MCPGetContextResponse(
            context_id=context_id,
            model=model,
            nodes=context["nodes"],
            metadata=context["metadata"]
        )
    
    except HTTPException:
        raise
    
    except Exception as e:
        logger.error(f"Error getting context: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get context: {str(e)}")

@app.delete("/v1/contexts/{context_id}", response_model=MCPDeleteContextResponse)
async def delete_context(context_id: str):
    """Delete a context by ID."""
    try:
        if context_id not in active_contexts:
            return MCPDeleteContextResponse(context_id=context_id, deleted=False)
        
        if active_contexts_lock.get(context_id, False):
            raise HTTPException(status_code=409, detail=f"Context {context_id} is currently being modified by another request")
        
        # Delete the context
        del active_contexts[context_id]
        if context_id in active_contexts_lock:
            del active_contexts_lock[context_id]
        
        return MCPDeleteContextResponse(context_id=context_id, deleted=True)
    
    except HTTPException:
        raise
    
    except Exception as e:
        logger.error(f"Error deleting context: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to delete context: {str(e)}")

@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {
        "status": "ok", 
        "gpu_available": torch.cuda.is_available(),
        "active_contexts": len(active_contexts),
        "loaded_models": len(loaded_models)
    }

@app.get("/models")
async def list_models():
    """List all loaded models."""
    return {
        "loaded_models": list(loaded_models.keys()),
        "models_info": {
            name: {
                "loaded_at": data["loaded_at"],
                "device": data["device"],
                "config": data["config"]
            } for name, data in loaded_models.items()
        }
    }

def main():
    """Main entry point for the MCP server."""
    parser = argparse.ArgumentParser(description="Model Context Protocol Server")
    parser.add_argument("--host", default="0.0.0.0", help="Host to bind to")
    parser.add_argument("--port", type=int, default=8001, help="Port to bind to")
    parser.add_argument("--preload", help="Model to preload on startup")
    args = parser.parse_args()
    
    # Preload model if specified
    if args.preload:
        try:
            load_model(args.preload)
        except Exception as e:
            logger.error(f"Failed to preload model {args.preload}: {e}")
    
    # Start the server
    logger.info(f"Starting MCP server on {args.host}:{args.port}")
    uvicorn.run(app, host=args.host, port=args.port, log_level="info")

if __name__ == "__main__":
    main()
