import request from "supertest";
import { app, resetState } from "./app";

beforeEach(() => {
  resetState();
});

describe("GET /health", () => {
  it("returns ok status", async () => {
    const res = await request(app).get("/health");
    expect(res.status).toBe(200);
    expect(res.body.status).toBe("ok");
    expect(res.body.service).toBe("dashboard-api");
  });
});

describe("POST /alerts", () => {
  it("creates an alert", async () => {
    const res = await request(app).post("/alerts").send({
      service: "auth-service",
      message: "High CPU usage",
      severity: "warning",
    });
    expect(res.status).toBe(201);
    expect(res.body.id).toBe("alert-1");
    expect(res.body.acknowledged).toBe(false);
  });

  it("rejects missing fields", async () => {
    const res = await request(app).post("/alerts").send({ service: "svc" });
    expect(res.status).toBe(400);
  });

  it("rejects invalid severity", async () => {
    const res = await request(app).post("/alerts").send({
      service: "svc",
      message: "test",
      severity: "extreme",
    });
    expect(res.status).toBe(400);
  });
});

describe("GET /alerts", () => {
  it("returns all alerts", async () => {
    await request(app).post("/alerts").send({
      service: "svc-a",
      message: "test1",
      severity: "info",
    });
    await request(app).post("/alerts").send({
      service: "svc-b",
      message: "test2",
      severity: "critical",
    });

    const res = await request(app).get("/alerts");
    expect(res.status).toBe(200);
    expect(res.body.count).toBe(2);
  });

  it("filters by service", async () => {
    await request(app).post("/alerts").send({
      service: "svc-a",
      message: "test1",
      severity: "info",
    });
    await request(app).post("/alerts").send({
      service: "svc-b",
      message: "test2",
      severity: "critical",
    });

    const res = await request(app).get("/alerts?service=svc-a");
    expect(res.body.count).toBe(1);
  });
});

describe("PATCH /alerts/:id/acknowledge", () => {
  it("acknowledges an alert", async () => {
    await request(app).post("/alerts").send({
      service: "svc",
      message: "test",
      severity: "info",
    });

    const res = await request(app).patch("/alerts/alert-1/acknowledge");
    expect(res.status).toBe(200);
    expect(res.body.acknowledged).toBe(true);
  });

  it("returns 404 for unknown alert", async () => {
    const res = await request(app).patch("/alerts/unknown/acknowledge");
    expect(res.status).toBe(404);
  });
});

describe("GET /dashboard/summary", () => {
  it("returns summary", async () => {
    await request(app).post("/alerts").send({
      service: "svc",
      message: "t1",
      severity: "info",
    });
    await request(app).post("/alerts").send({
      service: "svc",
      message: "t2",
      severity: "critical",
    });
    await request(app).patch("/alerts/alert-1/acknowledge");

    const res = await request(app).get("/dashboard/summary");
    expect(res.status).toBe(200);
    expect(res.body.total).toBe(2);
    expect(res.body.unacknowledged).toBe(1);
    expect(res.body.bySeverity.info).toBe(1);
    expect(res.body.bySeverity.critical).toBe(1);
  });
});
