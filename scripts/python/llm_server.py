"""
LLM Server module that provides a simple HTTP server for LLM inference.
This can be used by the Go service to run Python model inference.
"""

import argparse
import json
import logging
import os
import sys
import time
from typing import Dict, List, Optional, Union, Any

import torch
import uvicorn
from fastapi import FastAPI, HTTPException, Request, Response
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from transformers import AutoModelForCausalLM, AutoTokenizer, pipeline

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger("llm-server")

# Create FastAPI app
app = FastAPI(title="LLM Inference Server")

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Global variables for loaded models
loaded_models = {}

class GenerateOptions(BaseModel):
    """Options for text generation."""
    max_tokens: int = Field(512, description="Maximum number of tokens to generate")
    temperature: float = Field(0.7, description="Sampling temperature")
    top_p: float = Field(0.9, description="Nucleus sampling parameter")
    top_k: int = Field(50, description="Top-k sampling parameter")
    stop_tokens: List[str] = Field([], description="Tokens at which to stop generation")

class GenerateRequest(BaseModel):
    """Request for text generation."""
    model: str = Field(..., description="Model name or path")
    prompt: str = Field(..., description="Input prompt")
    options: GenerateOptions = Field(default_factory=GenerateOptions, description="Generation options")

class GenerateResponse(BaseModel):
    """Response from text generation."""
    text: str = Field(..., description="Generated text")
    tokens_used: int = Field(..., description="Number of tokens used")
    tokens_total: int = Field(..., description="Total number of tokens (prompt + generated)")
    time_taken: float = Field(..., description="Time taken for generation in seconds")

class ModelInfoResponse(BaseModel):
    """Response with model information."""
    name: str = Field(..., description="Model name")
    provider: str = Field(..., description="Model provider")
    max_context_length: int = Field(..., description="Maximum context length")
    required_memory: int = Field(..., description="Required memory in bytes")
    requires_gpu: bool = Field(..., description="Whether the model requires a GPU")

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
    
    # Load the model with optimizations appropriate for the model size and type
    kwargs = {
        "trust_remote_code": True,  # Needed for Qwen and some DeepSeek models
        "device_map": device
    }
    
    # Model-specific optimizations
    if "int8" in model_name.lower():
        kwargs["load_in_8bit"] = True
    elif "int4" in model_name.lower():
        kwargs["load_in_4bit"] = True
        
    # Special handling for Qwen models
    if "qwen" in model_name.lower():
        kwargs["torch_dtype"] = torch.bfloat16 if torch.cuda.is_available() else torch.float32
        
    # Special handling for DeepSeek models
    if "deepseek" in model_name.lower():
        kwargs["torch_dtype"] = torch.bfloat16 if torch.cuda.is_available() else torch.float32
    
    start_time = time.time()
    
    try:
        model = AutoModelForCausalLM.from_pretrained(
            model_path,
            **kwargs
        )
        tokenizer = AutoTokenizer.from_pretrained(model_path, trust_remote_code=True)
        
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
            "loaded_at": time.time()
        }
        
        loaded_models[model_name] = model_data
        return model_data
    
    except Exception as e:
        logger.error(f"Error loading model {model_name}: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to load model: {str(e)}")

@app.post("/generate", response_model=GenerateResponse)
async def generate_text(request: GenerateRequest):
    """Generate text from a prompt."""
    try:
        model_data = load_model(request.model)
        
        start_time = time.time()
        
        # Apply generation options
        options = request.options
        gen_kwargs = {
            "max_length": len(model_data["tokenizer"].encode(request.prompt)) + options.max_tokens,
            "temperature": options.temperature,
            "do_sample": options.temperature > 0,
            "top_p": options.top_p,
            "top_k": options.top_k,
            "num_return_sequences": 1
        }
        
        if options.stop_tokens:
            gen_kwargs["stopping_criteria"] = options.stop_tokens
        
        # Generate text
        result = model_data["pipeline"](
            request.prompt,
            **gen_kwargs
        )
        
        # Extract generated text
        generated_text = result[0]["generated_text"]
        
        # Remove the prompt from the beginning of the generated text
        if request.prompt in generated_text:
            generated_text = generated_text[len(request.prompt):]
        
        # Count tokens
        input_tokens = len(model_data["tokenizer"].encode(request.prompt))
        output_tokens = len(model_data["tokenizer"].encode(generated_text))
        total_tokens = input_tokens + output_tokens
        
        # Prepare response
        response = GenerateResponse(
            text=generated_text,
            tokens_used=output_tokens,
            tokens_total=total_tokens,
            time_taken=time.time() - start_time
        )
        
        logger.info(f"Generated {output_tokens} tokens in {response.time_taken:.2f} seconds")
        return response
    
    except Exception as e:
        logger.error(f"Error generating text: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to generate text: {str(e)}")

@app.get("/models/{model_name}", response_model=ModelInfoResponse)
async def get_model_info(model_name: str):
    """Get information about a model."""
    try:
        # If the model is loaded, we can get more accurate information
        if model_name in loaded_models:
            model_data = loaded_models[model_name]
            model = model_data["model"]
            
            # Estimate memory usage
            if hasattr(model, "get_memory_footprint"):
                memory_bytes = model.get_memory_footprint()
            else:
                # Rough estimate based on model parameters
                params = sum(p.numel() for p in model.parameters())
                # Assume 4 bytes per parameter for FP32, or 2 bytes for FP16
                bytes_per_param = 2 if model_data["device"] == "cuda" else 4
                memory_bytes = params * bytes_per_param
            
            return ModelInfoResponse(
                name=model_name,
                provider="local",
                max_context_length=model.config.max_position_embeddings if hasattr(model.config, "max_position_embeddings") else 2048,
                required_memory=memory_bytes,
                requires_gpu=model_data["device"] == "cuda"
            )
        
        # For models that aren't loaded, provide estimated information
        if "llama" in model_name.lower():
            size = "7b"
            if "13b" in model_name.lower():
                size = "13b"
            elif "70b" in model_name.lower():
                size = "70b"
            
            memory_map = {
                "7b": 7_000_000_000,
                "13b": 13_000_000_000,
                "70b": 70_000_000_000
            }
            
            context_length = 4096
            if "v2" in model_name.lower():
                context_length = 4096
            
            return ModelInfoResponse(
                name=model_name,
                provider="local",
                max_context_length=context_length,
                required_memory=memory_map.get(size, 7_000_000_000),
                requires_gpu=True
            )
        
        elif "mistral" in model_name.lower():
            return ModelInfoResponse(
                name=model_name,
                provider="local",
                max_context_length=8192,
                required_memory=8_000_000_000,
                requires_gpu=True
            )
        
        elif "qwen" in model_name.lower():
            size = "7b"
            if "7b" in model_name.lower():
                size = "7b"
            elif "14b" in model_name.lower():
                size = "14b"
            elif "72b" in model_name.lower():
                size = "72b"
                
            memory_map = {
                "7b": 7_000_000_000,
                "14b": 14_000_000_000,
                "72b": 72_000_000_000
            }
            
            return ModelInfoResponse(
                name=model_name,
                provider="Qwen",
                max_context_length=32768,  # Qwen3 supports 32k context
                required_memory=memory_map.get(size, 7_000_000_000),
                requires_gpu=True
            )
            
        elif "deepseek" in model_name.lower():
            size = "7b"
            if "7b" in model_name.lower():
                size = "7b"
            elif "33b" in model_name.lower():
                size = "33b"
                
            memory_map = {
                "7b": 7_000_000_000,
                "33b": 33_000_000_000
            }
            
            return ModelInfoResponse(
                name=model_name,
                provider="DeepSeek",
                max_context_length=16384,  # DeepSeek v3 supports 16k context
                required_memory=memory_map.get(size, 7_000_000_000),
                requires_gpu=True
            )
            
        # Generic fallback
        return ModelInfoResponse(
            name=model_name,
            provider="local",
            max_context_length=4096,
            required_memory=8_000_000_000,
            requires_gpu=True
        )
    
    except Exception as e:
        logger.error(f"Error getting model info for {model_name}: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get model info: {str(e)}")

@app.get("/models")
async def list_models():
    """List all loaded models."""
    return {
        "loaded_models": list(loaded_models.keys()),
        "models_info": {
            name: {
                "loaded_at": data["loaded_at"],
                "device": data["device"]
            } for name, data in loaded_models.items()
        }
    }

@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "ok", "gpu_available": torch.cuda.is_available()}

def main():
    """Main entry point for the LLM server."""
    parser = argparse.ArgumentParser(description="LLM Inference Server")
    parser.add_argument("--host", default="0.0.0.0", help="Host to bind to")
    parser.add_argument("--port", type=int, default=8000, help="Port to bind to")
    parser.add_argument("--preload", help="Model to preload on startup")
    args = parser.parse_args()
    
    # Preload model if specified
    if args.preload:
        try:
            load_model(args.preload)
        except Exception as e:
            logger.error(f"Failed to preload model {args.preload}: {e}")
    
    # Start the server
    logger.info(f"Starting LLM server on {args.host}:{args.port}")
    uvicorn.run(app, host=args.host, port=args.port, log_level="info")

if __name__ == "__main__":
    main()
