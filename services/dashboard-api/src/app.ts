import express, { Request, Response } from "express";

const app = express();
app.use(express.json());

interface Alert {
  id: string;
  service: string;
  message: string;
  severity: "info" | "warning" | "critical";
  createdAt: string;
  acknowledged: boolean;
}

const alerts: Alert[] = [];
let alertCounter = 0;

const log = (level: string, msg: string): void => {
  const ts = new Date().toISOString();
  console.log(`${ts} [${level}] dashboard-api: ${msg}`);
};

app.get("/health", (_req: Request, res: Response) => {
  res.json({ status: "ok", service: "dashboard-api" });
});

app.get("/alerts", (_req: Request, res: Response) => {
  const service = _req.query.service as string | undefined;
  let filtered = alerts;
  if (service) {
    filtered = alerts.filter((a) => a.service === service);
  }
  res.json({ count: filtered.length, alerts: filtered });
});

app.post("/alerts", (req: Request, res: Response) => {
  const { service, message, severity } = req.body;

  if (!service || !message || !severity) {
    log("WARN", "Alert creation with missing fields");
    res.status(400).json({ error: "service, message, and severity are required" });
    return;
  }

  const validSeverities = ["info", "warning", "critical"];
  if (!validSeverities.includes(severity)) {
    log("WARN", `Invalid severity: ${severity}`);
    res.status(400).json({
      error: `severity must be one of: ${validSeverities.join(", ")}`,
    });
    return;
  }

  alertCounter++;
  const alert: Alert = {
    id: `alert-${alertCounter}`,
    service,
    message,
    severity,
    createdAt: new Date().toISOString(),
    acknowledged: false,
  };

  alerts.push(alert);
  log("INFO", `Alert created: ${alert.id} for ${service} [${severity}]`);
  res.status(201).json(alert);
});

app.patch("/alerts/:id/acknowledge", (req: Request, res: Response) => {
  const alert = alerts.find((a) => a.id === req.params.id);
  if (!alert) {
    res.status(404).json({ error: "alert not found" });
    return;
  }

  alert.acknowledged = true;
  log("INFO", `Alert acknowledged: ${alert.id}`);
  res.json(alert);
});

app.get("/dashboard/summary", (_req: Request, res: Response) => {
  const total = alerts.length;
  const unacknowledged = alerts.filter((a) => !a.acknowledged).length;
  const bySeverity = {
    info: alerts.filter((a) => a.severity === "info").length,
    warning: alerts.filter((a) => a.severity === "warning").length,
    critical: alerts.filter((a) => a.severity === "critical").length,
  };

  res.json({ total, unacknowledged, bySeverity });
});

export { app, alerts, alertCounter };

export function resetState(): void {
  alerts.length = 0;
  alertCounter = 0;
}
