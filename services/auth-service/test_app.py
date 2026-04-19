import pytest
from app import app, users_db, tokens_db


@pytest.fixture
def client():
    app.config["TESTING"] = True
    users_db.clear()
    tokens_db.clear()
    with app.test_client() as client:
        yield client


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "ok"
    assert data["service"] == "auth-service"


def test_register_success(client):
    resp = client.post("/register", json={"username": "alice", "password": "secret"})
    assert resp.status_code == 201
    assert resp.get_json()["username"] == "alice"


def test_register_duplicate(client):
    client.post("/register", json={"username": "alice", "password": "secret"})
    resp = client.post("/register", json={"username": "alice", "password": "secret"})
    assert resp.status_code == 409


def test_register_missing_fields(client):
    resp = client.post("/register", json={"username": "alice"})
    assert resp.status_code == 400


def test_login_success(client):
    client.post("/register", json={"username": "alice", "password": "secret"})
    resp = client.post("/login", json={"username": "alice", "password": "secret"})
    assert resp.status_code == 200
    assert "token" in resp.get_json()


def test_login_wrong_password(client):
    client.post("/register", json={"username": "alice", "password": "secret"})
    resp = client.post("/login", json={"username": "alice", "password": "wrong"})
    assert resp.status_code == 401


def test_login_missing_fields(client):
    resp = client.post("/login", json={})
    assert resp.status_code == 400


def test_verify_valid_token(client):
    client.post("/register", json={"username": "alice", "password": "secret"})
    login_resp = client.post(
        "/login", json={"username": "alice", "password": "secret"}
    )
    token = login_resp.get_json()["token"]
    resp = client.post("/verify", json={"token": token})
    assert resp.status_code == 200
    assert resp.get_json()["valid"] is True


def test_verify_invalid_token(client):
    resp = client.post("/verify", json={"token": "invalid"})
    assert resp.status_code == 401
    assert resp.get_json()["valid"] is False


def test_verify_missing_token(client):
    resp = client.post("/verify", json={})
    assert resp.status_code == 400
