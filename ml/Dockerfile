FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt
COPY . ml/
RUN python ml/scripts/build_datasets.py && python ml/train.py --quick
EXPOSE 8000
CMD ["uvicorn", "ml.app:app", "--host", "0.0.0.0", "--port", "8000"]
