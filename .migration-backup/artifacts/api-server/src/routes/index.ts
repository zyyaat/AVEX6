import { Router, type IRouter } from "express";
import type { Request, Response } from "express";
import healthRouter from "./health";

const router: IRouter = Router();

router.use(healthRouter);

// ---------------------------------------------------------------------------
// Catch-all proxy → Go backend
// All /api/* routes not handled above are forwarded to BACKEND_URL.
// ---------------------------------------------------------------------------
const BACKEND_URL = process.env["BACKEND_URL"] || "http://localhost:8080";

router.all("/{*path}", async (req: Request, res: Response) => {
  const qs = req.url.includes("?") ? "?" + req.url.split("?").slice(1).join("?") : "";
  const url = `${BACKEND_URL}/api${req.path}${qs}`;

  try {
    const headers: Record<string, string> = {};
    for (const [key, value] of Object.entries(req.headers)) {
      if (
        key !== "host" &&
        key !== "connection" &&
        key !== "transfer-encoding" &&
        typeof value === "string"
      ) {
        headers[key] = value;
      }
    }

    const isBodyMethod = !["GET", "HEAD", "OPTIONS"].includes(req.method.toUpperCase());
    let body: string | undefined;
    if (isBodyMethod && req.body && Object.keys(req.body).length > 0) {
      body = JSON.stringify(req.body);
      headers["content-type"] = "application/json";
    }

    const backendRes = await fetch(url, { method: req.method, headers, body });

    res.status(backendRes.status);
    backendRes.headers.forEach((value, key) => {
      if (!["transfer-encoding", "connection", "keep-alive"].includes(key.toLowerCase())) {
        res.setHeader(key, value);
      }
    });

    const data = await backendRes.text();
    res.send(data);
  } catch {
    res.status(502).json({ error: "Backend unavailable" });
  }
});

export default router;
