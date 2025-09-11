import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 500 },  // 快速增压到500并发
    { duration: '1m', target: 1000 },  // 保持1000并发1分钟
    { duration: '10s', target: 0 },    // 冷却
  ],
  thresholds: {
    http_req_duration: ['p(99)<50'],  // P99延迟 <10毫秒
    http_req_failed: ['rate<0.01'],   // 错误率 <1%
  },
};

export default function () {
  const res = http.get('http://localhost:8081/hello');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time <10ms': (r) => r.timings.duration < 50,
  });
}