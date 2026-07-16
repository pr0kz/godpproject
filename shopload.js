import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

const errorRate = new Rate("application_errors");
const requestDuration = new Trend("shops_request_duration", true);

export const options = {
  scenarios: {
    shops_read: {
      executor: "ramping-vus",
      exec: "shopsRead",
      stages: [
        { duration: "10s", target: 20 },
        { duration: "30s", target: 100 },
        { duration: "20s", target: 200 },
        { duration: "10s", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500", "p(99)<1000"],
    application_errors: ["rate<0.01"],
  },
};

export function shopsRead() {
  const response = http.get(
    "http://localhost:8081/shops?page=1&page_size=10"
  );

  const success = check(response, {
    "status is 200": (r) => r.status === 200,
    "response is JSON": (r) =>
      String(r.headers["Content-Type"] || "").includes("application/json"),
    "response contains shops": (r) => {
      try {
        return Array.isArray(r.json("shops"));
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!success);
  requestDuration.add(response.timings.duration);

  sleep(0.1);
}