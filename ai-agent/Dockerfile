FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Drop privileges: run as a non-root user.
RUN useradd -m -u 1000 aiagent
USER aiagent

EXPOSE 8090

CMD ["uvicorn", "app:api", "--host", "0.0.0.0", "--port", "8090"]
