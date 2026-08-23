import http from 'k6/http';
import exec from 'k6/execution';
import { check } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

const datasetPath = __ENV.PHASE2_DATASET || './.generated/phase2-users.json';
const data = new SharedArray('phase2 buyers', () => [JSON.parse(open(datasetPath))])[0];
const baseURL = __ENV.BASE_URL || 'http://127.0.0.1:8080';
const vus = Number(__ENV.VUS || Math.min(200, data.tokens.length));
const iterations = Number(__ENV.ITERATIONS || data.tokens.length);

const successfulOrders = new Counter('seckill_success');
const soldOutResponses = new Counter('seckill_sold_out');
const busyResponses = new Counter('seckill_busy');
const unexpectedResponses = new Counter('seckill_unexpected');
const optimisticRetries = new Trend('seckill_optimistic_retries', true);

export const options = {
  summaryTrendStats: ['min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max', 'avg'],
  scenarios: {
    seckill: {
      executor: 'shared-iterations',
      vus,
      iterations,
      maxDuration: __ENV.MAX_DURATION || '45s',
      gracefulStop: '0s',
    },
  },
  thresholds: {
    checks: ['rate==1'],
    seckill_unexpected: ['count==0'],
    http_req_duration: ['p(95)<3000', 'p(99)<5000'],
  },
  discardResponseBodies: false,
};

export default function () {
  const index = exec.scenario.iterationInTest;
  const response = http.post(
    `${baseURL}/api/v1/activities/${data.activity_id}/seckill`,
    JSON.stringify({ quantity: 1 }),
    {
      headers: {
        Authorization: `Bearer ${data.tokens[index]}`,
        'Content-Type': 'application/json',
        'X-Request-ID': `phase2-${__ENV.STRATEGY || 'unknown'}-${index}`,
      },
      responseCallback: http.expectedStatuses(201, 409, 503),
      tags: { strategy: __ENV.STRATEGY || 'unknown' },
    },
  );

  let code = '';
  let retries = 0;
  try {
    const body = response.json();
    code = body.code || '';
    retries = body.data?.concurrency_retries || 0;
  } catch (_) {
    // The check below turns malformed responses into an explicit test failure.
  }

  if (response.status === 201) successfulOrders.add(1);
  else if (response.status === 409 && code === 'out_of_stock') soldOutResponses.add(1);
  else if (response.status === 503 && code === 'concurrency_busy') busyResponses.add(1);
  else unexpectedResponses.add(1);
  optimisticRetries.add(retries);

  check(response, {
    'response is success, sold out, or controlled busy': () =>
      response.status === 201 ||
      (response.status === 409 && code === 'out_of_stock') ||
      (response.status === 503 && code === 'concurrency_busy'),
  });
}

export function handleSummary(summary) {
  const path = __ENV.K6_SUMMARY_PATH || './.generated/phase2-summary.json';
  return {
    stdout: `${textSummary(summary)}\n`,
    [path]: JSON.stringify(summary, null, 2),
  };
}

function textSummary(summary) {
  const metric = (name, key = 'count') => summary.metrics[name]?.values?.[key] ?? 0;
  return [
    `strategy=${__ENV.STRATEGY || 'unknown'}`,
    `requests=${metric('http_reqs')}`,
    `success=${metric('seckill_success')}`,
    `sold_out=${metric('seckill_sold_out')}`,
    `busy=${metric('seckill_busy')}`,
    `qps=${metric('http_reqs', 'rate').toFixed(2)}`,
    `p50_ms=${metric('http_req_duration', 'med').toFixed(2)}`,
    `p95_ms=${metric('http_req_duration', 'p(95)').toFixed(2)}`,
    `p99_ms=${metric('http_req_duration', 'p(99)').toFixed(2)}`,
  ].join(' ');
}
