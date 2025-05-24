import torch
from transformers import AutoModelForCausalLM, AutoTokenizer
from flask import Flask, request, jsonify
import logging
import os
import argparse
from typing import Dict, Any, List, Optional
import time

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# 全局变量存储模型和tokenizer
MODEL = None
TOKENIZER = None
DEVICE = None
MAX_LENGTH = 2048
DEFAULT_TEMPERATURE = 0.7

def load_model(model_path: str, device: str = "cuda") -> None:
    """
    加载DeepSeek V3 7B模型
    
    Args:
        model_path: 模型路径
        device: 设备类型 (cuda或cpu)
    """
    global MODEL, TOKENIZER, DEVICE
    
    DEVICE = device if torch.cuda.is_available() and device == "cuda" else "cpu"
    logger.info(f"正在加载DeepSeek V3 7B模型到{DEVICE}...")
    
    try:
        # 加载tokenizer
        TOKENIZER = AutoTokenizer.from_pretrained(model_path, trust_remote_code=True)
        
        # 加载模型
        MODEL = AutoModelForCausalLM.from_pretrained(
            model_path,
            torch_dtype=torch.float16 if DEVICE == "cuda" else torch.float32,
            device_map=DEVICE,
            trust_remote_code=True
        )
        
        logger.info("模型加载完成!")
    except Exception as e:
        logger.error(f"模型加载失败: {str(e)}")
        raise

@app.route('/health', methods=['GET'])
def health_check():
    """健康检查接口"""
    if MODEL is not None and TOKENIZER is not None:
        return jsonify({"status": "ok", "model": "DeepSeek V3 7B"}), 200
    return jsonify({"status": "error", "message": "模型未加载"}), 500

@app.route('/v1/generate', methods=['POST'])
def generate():
    """文本生成接口"""
    start_time = time.time()
    
    if MODEL is None or TOKENIZER is None:
        return jsonify({"error": "模型未加载"}), 503
    
    # 获取请求数据
    data = request.json
    if not data:
        return jsonify({"error": "无效的请求体"}), 400
    
    # 提取参数
    prompt = data.get("prompt", "")
    if not prompt:
        return jsonify({"error": "缺少prompt参数"}), 400
    
    max_length = data.get("max_length", MAX_LENGTH)
    temperature = data.get("temperature", DEFAULT_TEMPERATURE)
    top_p = data.get("top_p", 0.95)
    stop_sequences = data.get("stop", [])
    
    try:
        # 处理输入
        input_ids = TOKENIZER.encode(prompt, return_tensors="pt").to(DEVICE)
        
        # 生成文本
        with torch.no_grad():
            output = MODEL.generate(
                input_ids,
                max_length=max_length,
                do_sample=temperature > 0,
                temperature=temperature,
                top_p=top_p,
                pad_token_id=TOKENIZER.eos_token_id
            )
        
        # 解码输出
        generated_text = TOKENIZER.decode(output[0], skip_special_tokens=True)
        
        # 如果设置了停止序列，在第一个停止序列处截断
        if stop_sequences:
            for stop_seq in stop_sequences:
                if stop_seq in generated_text:
                    generated_text = generated_text[:generated_text.find(stop_seq)]
        
        # 去除原始提示
        if generated_text.startswith(prompt):
            generated_text = generated_text[len(prompt):]
        
        # 计算处理时间
        processing_time = time.time() - start_time
        
        # 返回结果
        return jsonify({
            "text": generated_text,
            "processing_time": processing_time,
            "model": "DeepSeek V3 7B",
        })
    
    except Exception as e:
        logger.error(f"生成过程中出错: {str(e)}")
        return jsonify({"error": f"生成过程中出错: {str(e)}"}), 500

@app.route('/v1/chat/completions', methods=['POST'])
def chat():
    """聊天完成接口 (OpenAI API 兼容)"""
    start_time = time.time()
    
    if MODEL is None or TOKENIZER is None:
        return jsonify({"error": "模型未加载"}), 503
    
    # 获取请求数据
    data = request.json
    if not data:
        return jsonify({"error": "无效的请求体"}), 400
    
    # 提取参数
    messages = data.get("messages", [])
    if not messages:
        return jsonify({"error": "缺少messages参数"}), 400
    
    max_tokens = data.get("max_tokens", MAX_LENGTH)
    temperature = data.get("temperature", DEFAULT_TEMPERATURE)
    top_p = data.get("top_p", 0.95)
    stop = data.get("stop", [])
    
    try:
        # 构造提示
        prompt = ""
        for message in messages:
            role = message.get("role", "user")
            content = message.get("content", "")
            
            if role == "system":
                prompt += f"<|system|>\n{content}\n"
            elif role == "user":
                prompt += f"<|user|>\n{content}\n"
            elif role == "assistant":
                prompt += f"<|assistant|>\n{content}\n"
        
        # 添加最后的assistant标记
        prompt += "<|assistant|>\n"
        
        # 处理输入
        input_ids = TOKENIZER.encode(prompt, return_tensors="pt").to(DEVICE)
        
        # 生成文本
        with torch.no_grad():
            output = MODEL.generate(
                input_ids,
                max_length=input_ids.shape[1] + max_tokens,
                do_sample=temperature > 0,
                temperature=temperature,
                top_p=top_p,
                pad_token_id=TOKENIZER.eos_token_id
            )
        
        # 解码输出
        generated_text = TOKENIZER.decode(output[0], skip_special_tokens=True)
        
        # 提取生成的回复
        if "<|assistant|>" in generated_text:
            parts = generated_text.split("<|assistant|>")
            response_text = parts[-1].strip()
        else:
            response_text = generated_text[len(prompt):].strip()
        
        # 如果设置了停止序列，在第一个停止序列处截断
        if stop:
            for stop_seq in stop:
                if stop_seq in response_text:
                    response_text = response_text[:response_text.find(stop_seq)]
        
        # 计算处理时间
        processing_time = time.time() - start_time
        
        # 构造OpenAI兼容的响应格式
        response = {
            "id": f"chatcmpl-{int(time.time())}",
            "object": "chat.completion",
            "created": int(time.time()),
            "model": "DeepSeek V3 7B",
            "choices": [
                {
                    "index": 0,
                    "message": {
                        "role": "assistant",
                        "content": response_text
                    },
                    "finish_reason": "stop"
                }
            ],
            "usage": {
                "prompt_tokens": len(input_ids[0]),
                "completion_tokens": len(TOKENIZER.encode(response_text)),
                "total_tokens": len(input_ids[0]) + len(TOKENIZER.encode(response_text))
            }
        }
        
        return jsonify(response)
    
    except Exception as e:
        logger.error(f"聊天过程中出错: {str(e)}")
        return jsonify({"error": f"聊天过程中出错: {str(e)}"}), 500

def main():
    """主函数，解析参数并启动服务"""
    parser = argparse.ArgumentParser(description='DeepSeek V3 7B 模型服务')
    parser.add_argument('--model_path', type=str, default="deepseek-ai/deepseek-v3-7b", 
                        help='模型路径或Hugging Face模型ID')
    parser.add_argument('--port', type=int, default=5000, help='服务端口')
    parser.add_argument('--host', type=str, default='0.0.0.0', help='服务地址')
    parser.add_argument('--device', type=str, default='cuda', choices=['cuda', 'cpu'], help='运行设备')
    
    args = parser.parse_args()
    
    # 加载模型
    load_model(args.model_path, args.device)
    
    # 启动服务
    app.run(host=args.host, port=args.port)

if __name__ == '__main__':
    main()
