import { app } from "./app";

const port = parseInt(process.env.DASHBOARD_API_PORT || "5003", 10);

app.listen(port, "0.0.0.0", () => {
  const ts = new Date().toISOString();
  console.log(`${ts} [INFO] dashboard-api: Starting on port ${port}`);
});
