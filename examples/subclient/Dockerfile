FROM python:3.7-alpine
COPY . /app
WORKDIR /app
RUN pip install flask flask_cors -i https://pypi.tuna.tsinghua.edu.cn/simple
EXPOSE 5000
CMD ["python", "app.py"]
