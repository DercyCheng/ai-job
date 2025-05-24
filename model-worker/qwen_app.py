import torch
from transformers import AutoModelForCausalLM, AutoTokenizer
from flask import Flask, request, jsonify, Response, stream_with_context
import logging
import os
import argparse
from typing import Dict, Any, List, Optional
import time
import json

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
MODEL_NAME = "Qwen3"
MAX_LENGTH = 4096
DEFAULT_TEMPERATURE = 0.7

def load_model(model_path: str, device: str = "cuda") -> None:
    """
    加载Qwen3模型
    
    Args:
        model_path: 模型路径
        device: 设备类型 (cuda或cpu)
    """
    global MODEL, TOKENIZER, DEVICE
    
    DEVICE = device if torch.cuda.is_available() and device == "cuda" else "cpu"
    logger.info(f"正在加载{MODEL_NAME}模型到{DEVICE}...")
    
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
        
        logger.info(f"{MODEL_NAME}模型加载完成!")
    except Exception as e:
        logger.error(f"模型加载失败: {str(e)}")
        raise

@app.route('/health', methods=['GET'])
def health_check():
    """健康检查接口"""
    if MODEL is not None and TOKENIZER is not None:
        return jsonify({"status": "ok", "model": MODEL_NAME}), 200
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
            "model": MODEL_NAME,
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
    stream = data.get("stream", False)
    
    try:
        # 构造Qwen3的聊天格式
        chat_format = []
        for message in messages:
            role = message.get("role", "user")
            content = message.get("content", "")
            
            # Qwen3使用不同的格式化方式
            if role == "system":
                chat_format.append({"role": "system", "content": content})
            elif role == "user":
                chat_format.append({"role": "user", "content": content})
            elif role == "assistant":
                chat_format.append({"role": "assistant", "content": content})
        
        # 检查是否为流式响应
        if stream:
            return stream_chat_response(chat_format, max_tokens, temperature, top_p, stop)
        
        # 使用Qwen3的聊天API
        response_text = ""
        generation_config = {
            "max_new_tokens": max_tokens,
            "do_sample": temperature > 0,
            "temperature": temperature,
            "top_p": top_p,
        }
        
        response = MODEL.chat(TOKENIZER, chat_format, generation_config=generation_config)
        response_text = response["response"]
        
        # 如果设置了停止序列，在第一个停止序列处截断
        if stop:
            for stop_seq in stop:
                if stop_seq in response_text:
                    response_text = response_text[:response_text.find(stop_seq)]
        
        # 计算处理时间
        processing_time = time.time() - start_time
        
        # 构造OpenAI兼容的响应格式
        api_response = {
            "id": f"chatcmpl-{int(time.time())}",
            "object": "chat.completion",
            "created": int(time.time()),
            "model": MODEL_NAME,
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
                "prompt_tokens": len(TOKENIZER.encode(str(chat_format))),
                "completion_tokens": len(TOKENIZER.encode(response_text)),
                "total_tokens": len(TOKENIZER.encode(str(chat_format))) + len(TOKENIZER.encode(response_text))
            }
        }
        
        return jsonify(api_response)
    
    except Exception as e:
        logger.error(f"聊天过程中出错: {str(e)}")
        return jsonify({"error": f"聊天过程中出错: {str(e)}"}), 500

def stream_chat_response(chat_format, max_tokens, temperature, top_p, stop_sequences):
    """流式返回聊天响应"""
    try:
        generation_config = {
            "max_new_tokens": max_tokens,
            "do_sample": temperature > 0,
            "temperature": temperature,
            "top_p": top_p,
        }

        # 创建流式响应生成器
        def generate():
            # 开始标记
            chunk_id = f"chatcmpl-{int(time.time())}"
            created = int(time.time())
            
            # 发送SSE格式的初始响应
            yield f"data: {json.dumps({'id': chunk_id, 'object': 'chat.completion.chunk', 'created': created, 'model': MODEL_NAME, 'choices': [{'index': 0, 'delta': {'role': 'assistant'}, 'finish_reason': None}]})}\n\n"
            
            # 使用Qwen的流式API (实现会根据Qwen的具体接口而定)
            streamer = MODEL.chat_stream(TOKENIZER, chat_format, generation_config=generation_config)

            collected_text = ""
            for response in streamer:
                # 获取当前chunk的内容
                chunk_content = response.get("delta", "")
                collected_text += chunk_content
                
                # 检查停止序列
                finish_reason = None
                if stop_sequences:
                    for stop_seq in stop_sequences:
                        if stop_seq in collected_text:
                            # 找到停止序列，截断文本
                            chunk_content = chunk_content[:chunk_content.find(stop_seq)]
                            finish_reason = "stop"
                            break
                
                # 构造SSE格式的chunk
                chunk = {
                    "id": chunk_id,
                    "object": "chat.completion.chunk",
                    "created": created,
                    "model": MODEL_NAME,
                    "choices": [
                        {
                            "index": 0,
                            "delta": {"content": chunk_content},
                            "finish_reason": finish_reason
                        }
                    ]
                }
                
                yield f"data: {json.dumps(chunk)}\n\n"
                
                # 如果找到停止序列，停止生成
                if finish_reason:
                    break
            
            # 发送完成标记
            yield f"data: {json.dumps({'id': chunk_id, 'object': 'chat.completion.chunk', 'created': created, 'model': MODEL_NAME, 'choices': [{'index': 0, 'delta': {}, 'finish_reason': 'stop'}]})}\n\n"
            yield "data: [DONE]\n\n"
        
        # 返回流式响应
        return Response(stream_with_context(generate()), mimetype='text/event-stream')
    
    except Exception as e:
        logger.error(f"流式生成过程中出错: {str(e)}")
        # 返回错误响应
        return jsonify({"error": f"流式生成过程中出错: {str(e)}"}), 500

def main():
    """主函数，解析参数并启动服务"""
    parser = argparse.ArgumentParser(description='Qwen3 模型服务')
    parser.add_argument('--model_path', type=str, default="Qwen/Qwen1.5-7B-Chat", 
                        help='模型路径或Hugging Face模型ID')
    parser.add_argument('--port', type=int, default=5001, help='服务端口')
    parser.add_argument('--host', type=str, default='0.0.0.0', help='服务地址')
    parser.add_argument('--device', type=str, default='cuda', choices=['cuda', 'cpu'], help='运行设备')
    
    args = parser.parse_args()
    
    # 加载模型
    load_model(args.model_path, args.device)
    
    # 启动服务
    app.run(host=args.host, port=args.port)

if __name__ == '__main__':
    main()
