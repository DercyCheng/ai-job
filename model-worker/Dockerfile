FROM python:3.10-slim

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 复制需要的文件
COPY requirements.txt ./

# 安装Python依赖
RUN pip install --no-cache-dir -r requirements.txt

# 复制应用代码
COPY app.py ./

# 设置环境变量
ENV MODEL_PATH="deepseek-ai/deepseek-v3-7b"
ENV PORT=5000
ENV HOST="0.0.0.0"
ENV DEVICE="cuda"

# 暴露端口
EXPOSE 5000

# 启动服务
CMD ["python", "app.py", "--model_path", "${MODEL_PATH}", "--port", "5000", "--host", "0.0.0.0", "--device", "${DEVICE}"]
