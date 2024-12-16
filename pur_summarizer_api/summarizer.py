from flask import Flask, request, jsonify
from dotenv import load_dotenv
import requests
import os

app = Flask(__name__)

load_dotenv()

MEANINGCLOUD_API_KEY = os.getenv("API_KEY")
MEANINGCLOUD_ENDPOINT = "https://api.meaningcloud.com/summarization-1.0"

@app.route('/summarize', methods=['POST'])
def summarize():
    data = request.json
    if not data or 'url' not in data:
        return jsonify({"error": "URL not provided"}), 400

    article_url = data['url']

    payload = {
        'key': MEANINGCLOUD_API_KEY,
        'url': article_url,
        'sentences': 10
    }

    try:
        response = requests.post(MEANINGCLOUD_ENDPOINT, data=payload)
        response.raise_for_status()
        result = response.json()

        if result.get("status", {}).get("code") == 0:
            return jsonify({
                "summary": result.get("summary", "No summary available"),
                "status": "success"
            })
        else:
            return jsonify({
                "error": "MeaningCloud API returned an error",
                "details": result.get("status", {}).get("msg", "Unknown error")
            }), 500

    except requests.exceptions.RequestException as e:
        return jsonify({"error": "Request to MeaningCloud failed", "details": str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True)