"""Entry point for uvicorn:  uvicorn app.server:app --port 8000"""
from .config import load_settings
from .main import create_app

app = create_app(load_settings())
