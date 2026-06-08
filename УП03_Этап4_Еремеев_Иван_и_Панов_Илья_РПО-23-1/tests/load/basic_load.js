// LandMark - basic k6 load test (stage 4).
// Run: k6 run tests/load/basic_load.js
// Optional: BASE=http://localhost:8080 k6 run tests/load/basic_load.js
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE = __ENV.BASE || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '10s', target: 10 }, // ramp up to 10 virtual users
    { duration: '20s', target: 10 }, // hold
    { duration: '5s', target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],          // <1% errors
    http_req_duration: ['p(95)<1000'],       // 95% of requests < 1s
  },
};

export default function () {
  const health = http.get(`${BASE}/healthz`);
  check(health, {
    'health status is 200': (r) => r.status === 200,
    'health < 1000ms': (r) => r.timings.duration < 1000,
  });

  const places = http.get(`${BASE}/v1/places`);
  check(places, {
    'places status is 200': (r) => r.status === 200,
    'places returns body': (r) => r.body && r.body.length > 0,
  });

  sleep(1);
}
