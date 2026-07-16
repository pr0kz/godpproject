import http from "k6/http";
import { check } from "k6";
import { Counter, Rate } from "k6/metrics";

const accepted = new Counter("seckill_accepted");
const rejected = new Counter("seckill_rejected");
const unexpected = new Counter("seckill_unexpected");
const applicationErrors = new Rate("application_errors");

const BASE_URL = __ENV.BASE_URL || "http://localhost:8081";
const USER_COUNT = Number(__ENV.USER_COUNT || 200);
const COUPON_STOCK = Number(__ENV.COUPON_STOCK || 50);

export const options = {
  setupTimeout: __ENV.SETUP_TIMEOUT || "10m",
  scenarios: {
    seckill: {
      executor: "per-vu-iterations",
      exec: "seckill",
      vus: USER_COUNT,
      iterations: 1,
      maxDuration: __ENV.MAX_DURATION || "5m",
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<1000", "p(99)<2000"],
    http_req_failed: ["rate<0.80"],
    application_errors: ["rate<0.01"],
  },
};

function jsonHeaders(token) {
  const headers = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  return { headers };
}

function registerAndLogin(username) {
  const password = "Concurrent123!";
  const email = `${username}@example.com`;

  const registerResponse = http.post(
    `${BASE_URL}/register`,
    JSON.stringify({
      username,
      email,
      password,
    }),
    jsonHeaders()
  );

  if (registerResponse.status !== 201 && registerResponse.status !== 400) {
    throw new Error(
      `register ${username} failed: ${registerResponse.status} ${registerResponse.body}`
    );
  }

  const loginResponse = http.post(
    `${BASE_URL}/login`,
    JSON.stringify({
      username,
      password,
    }),
    jsonHeaders()
  );

  if (loginResponse.status !== 200) {
    throw new Error(
      `login ${username} failed: ${loginResponse.status} ${loginResponse.body}`
    );
  }

  const token = loginResponse.json("token");
  if (!token) {
    throw new Error(`login ${username} did not return a token`);
  }

  return token;
}

export function setup() {
  const runID = `${Date.now()}-${Math.floor(Math.random() * 100000)}`;
  const tokens = [];

  for (let index = 0; index < USER_COUNT; index += 1) {
    const username = `load-${runID}-${index}`;
    tokens.push(registerAndLogin(username));
  }

  const creatorToken = tokens[0];
  const now = Math.floor(Date.now() / 1000);

  const couponResponse = http.post(
    `${BASE_URL}/coupons`,
    JSON.stringify({
      title: `Concurrent coupon ${runID}`,
      stock: COUPON_STOCK,
      discount: 9.9,
      begin_time: now - 10,
      end_time: now + 600,
    }),
    jsonHeaders(creatorToken)
  );

  if (couponResponse.status !== 201) {
    throw new Error(
      `create coupon failed: ${couponResponse.status} ${couponResponse.body}`
    );
  }

  const couponID = couponResponse.json("coupon.id");
  if (!couponID) {
    throw new Error("create coupon did not return coupon.id");
  }

  console.log(
    `Created coupon ${couponID}, stock=${COUPON_STOCK}, users=${USER_COUNT}`
  );

  return {
    couponID,
    tokens,
  };
}

export function seckill(data) {
  const token = data.tokens[__VU - 1];

  const response = http.post(
    `${BASE_URL}/seckill/${data.couponID}`,
    null,
    jsonHeaders(token)
  );

  let errorMessage = "";
  try {
    errorMessage = String(response.json("error") || "");
  } catch {
    errorMessage = "";
  }

  const expectedRejection =
    response.status === 400 &&
    (errorMessage === "sold out" ||
      errorMessage === "you have already purchased this coupon");
  const expected = response.status === 202 || expectedRejection;

  check(response, {
    "response is accepted or expected rejection": () => expected,
  });

  applicationErrors.add(!expected);

  if (response.status === 202) {
    accepted.add(1);
  } else if (expectedRejection) {
    rejected.add(1);
  } else {
    unexpected.add(1);
    console.error(
      `VU ${__VU}: unexpected ${response.status}: ${response.body}`
    );
  }
}