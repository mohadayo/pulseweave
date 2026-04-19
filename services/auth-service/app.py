import logging
import os
import hashlib
import secrets
from flask import Flask, jsonify, request

app = Flask(__name__)

logging.basicConfig(
    level=os.environ.get("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("auth-service")

users_db: dict[str, dict] = {}
tokens_db: dict[str, str] = {}


def hash_password(password: str) -> str:
    return hashlib.sha256(password.encode()).hexdigest()


@app.route("/health")
def health():
    return jsonify({"status": "ok", "service": "auth-service"})


@app.route("/register", methods=["POST"])
def register():
    data = request.get_json()
    if not data or "username" not in data or "password" not in data:
        logger.warning("Registration attempt with missing fields")
        return jsonify({"error": "username and password are required"}), 400

    username = data["username"]
    if username in users_db:
        logger.warning("Registration attempt for existing user: %s", username)
        return jsonify({"error": "user already exists"}), 409

    users_db[username] = {
        "username": username,
        "password_hash": hash_password(data["password"]),
    }
    logger.info("User registered: %s", username)
    return jsonify({"message": "user registered", "username": username}), 201


@app.route("/login", methods=["POST"])
def login():
    data = request.get_json()
    if not data or "username" not in data or "password" not in data:
        logger.warning("Login attempt with missing fields")
        return jsonify({"error": "username and password are required"}), 400

    username = data["username"]
    user = users_db.get(username)
    if not user or user["password_hash"] != hash_password(data["password"]):
        logger.warning("Failed login attempt for: %s", username)
        return jsonify({"error": "invalid credentials"}), 401

    token = secrets.token_hex(32)
    tokens_db[token] = username
    logger.info("User logged in: %s", username)
    return jsonify({"token": token, "username": username})


@app.route("/verify", methods=["POST"])
def verify():
    data = request.get_json()
    if not data or "token" not in data:
        return jsonify({"error": "token is required"}), 400

    token = data["token"]
    username = tokens_db.get(token)
    if not username:
        return jsonify({"valid": False}), 401

    return jsonify({"valid": True, "username": username})


if __name__ == "__main__":
    port = int(os.environ.get("AUTH_SERVICE_PORT", 5001))
    logger.info("Starting auth-service on port %d", port)
    app.run(host="0.0.0.0", port=port)
