import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const BASE_URL = 'http:


export const latency = new Trend('http_req_duration_ms');
export const successRate = new Rate('successful_requests');

export const options = {
  vus: 5,
  duration: '60s',
  thresholds: {
    http_req_duration: ['p(95)<300'],
    successful_requests: ['rate>0.999'],
  },
};

const TEAM_NAME = 'team_k6';
const NEW_TEAM_NAME = 'team_k6_new';

const USERS = [
  { user_id: 'k6_u1', username: 'k6_user_1', is_active: true },
  { user_id: 'k6_u2', username: 'k6_user_2', is_active: true },
  { user_id: 'k6_u3', username: 'k6_user_3', is_active: true },
];

let prCounter = 0;

function record(res, ok) {
  latency.add(res.timings.duration);
  successRate.add(ok);
}

export default function () {
  
  {
    const res = http.get(`${BASE_URL}/health`);
    const ok = check(res, {
      'health status is 200': (r) => r.status === 200,
    });
    record(res, ok);
  }

  
  {
    const payload = JSON.stringify({
      team_name: TEAM_NAME,
      members: USERS,
    });

    const res = http.post(`${BASE_URL}/team/add`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const ok = check(res, {
      'team/add 201 or 200/400': (r) =>
        r.status === 201 || r.status === 200 || r.status === 400,
    });
    record(res, ok);
  }

  
  {
    const payload = JSON.stringify({
      team_name: NEW_TEAM_NAME,
      members: [
        { user_id: 'k6_new_1', username: 'k6_new_user_1', is_active: true },
        { user_id: 'k6_new_2', username: 'k6_new_user_2', is_active: true },
      ],
    });

    const res = http.post(`${BASE_URL}/team/add`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const ok = check(res, {
      'team/add new 201 or 200/400': (r) =>
        r.status === 201 || r.status === 200 || r.status === 400,
    });
    record(res, ok);
  }

  
  {
    const res = http.get(
      `${BASE_URL}/team/get?team_name=${encodeURIComponent(TEAM_NAME)}`
    );
    const ok = check(res, {
      'team/get 200': (r) => r.status === 200,
    });
    record(res, ok);
  }

  
  {
    const payload = JSON.stringify({
      user_id: USERS[1].user_id,
      is_active: false,
    });

    const res = http.post(`${BASE_URL}/users/setIsActive`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const ok = check(res, {
      'setIsActive 200 or 404': (r) => r.status === 200 || r.status === 404,
    });
    record(res, ok);
  }

  
  const prId = `k6-pr-${__VU}-${prCounter++}`;
  {
    const payload = JSON.stringify({
      pull_request_id: prId,
      pull_request_name: `k6 load test PR ${prId}`,
      author_id: USERS[0].user_id,
    });

    const res = http.post(`${BASE_URL}/pullRequest/create`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const ok = check(res, {
      'pr/create 201 or 404/409': (r) =>
        r.status === 201 || r.status === 404 || r.status === 409,
    });
    record(res, ok);
  }

  
  {
    const res = http.get(
      `${BASE_URL}/users/getReview?user_id=${encodeURIComponent(
        USERS[1].user_id,
      )}`,
    );

    const ok = check(res, {
      'users/getReview 200 or 404': (r) => r.status === 200 || r.status === 404,
    });
    record(res, ok);
  }

  
  {
    const res = http.get(`${BASE_URL}/users/stats`);
    const ok = check(res, {
      'users/stats 200': (r) => r.status === 200,
    });
    record(res, ok);
  }

  
  {
    const res = http.get(`${BASE_URL}/pullRequest/stats`);
    const ok = check(res, {
      'pullRequest/stats 200': (r) => r.status === 200,
    });
    record(res, ok);
  }

  
  {
    const payload = JSON.stringify({
      from_team_name: TEAM_NAME,
      to_team_name: NEW_TEAM_NAME,
    });

    const res = http.post(`${BASE_URL}/team/deactivate`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const ok = check(res, {
      'team/deactivate 200 or 404/409': (r) =>
        r.status === 200 || r.status === 404 || r.status === 409,
    });
    record(res, ok);
  }

  sleep(1);
}
