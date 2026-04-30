import os
from dotenv import load_dotenv
import google.generativeai as genai

load_dotenv(dotenv_path="../.env")
api_key = os.getenv("GEMINI_API_KEY")

if not api_key:
    print("Error: GEMINI_API_KEY not found in .env")
else:
    genai.configure(api_key=api_key)
    print("--- Available Gemini Models ---")
    try:
        for m in genai.list_models():
            if 'generateContent' in m.supported_generation_methods:
                print(f"Model Name: {m.name}")
    except Exception as e:
        print(f"Error listing models: {e}")
