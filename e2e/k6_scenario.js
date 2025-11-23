import http from 'k6/http';
import { check, sleep } from 'k6';

// const BASE_URL = 'http://localhost:8080';
const BASE_URL = __ENV.BASE_URL;
export const options = {
  vus: 1,
  iterations: 1,
};

const TEAM_NAME = 'team_k6';
const NEW_TEAM_NAME = 'team_k6_new';

const USERS = [
  { user_id: 'k6_u1', username: 'k6_user_1', is_active: true },
  { user_id: 'k6_u2', username: 'k6_user_2', is_active: true },
  { user_id: 'k6_u3', username: 'k6_user_3', is_active: true },
];

let prCounter = 0;


export default function () {
  
  {
    const res = http.get(`${BASE_URL}/health`);
    const ok = check(res, {
      'health 200': (r) => r.status === 200 && r.body === 'OK',
    });
    
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
      'team/add src ok': (r) => [200, 201, 400].includes(r.status),
    });
    
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
      'team/add dst ok': (r) => [200, 201, 400].includes(r.status),
    });
    
  }

  
  {
    const res = http.get(
      `${BASE_URL}/team/get?team_name=${encodeURIComponent(TEAM_NAME)}`
    );
    const ok = check(res, {
      'team/get 200 and has members': (r) => {
        if (r.status !== 200) return false;
        const body = r.json();
        return body.team_name === TEAM_NAME && body.members.length >= 1;
      },
    });
    
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
      'setIsActive 200': (r) => r.status === 200,
    });
    
  }

  
  const prId = `k6-pr-${__VU}-${prCounter++}`;
  {
    const payload = JSON.stringify({
      pull_request_id: prId,
      pull_request_name: `k6 e2e PR ${prId}`,
      author_id: USERS[0].user_id,
    });

    const res = http.post(`${BASE_URL}/pullRequest/create`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const ok = check(res, {
      'pr/create 201': (r) => r.status === 201,
    });
    
  }

  
  {
    const res = http.get(
      `${BASE_URL}/users/getReview?user_id=${encodeURIComponent(
        USERS[1].user_id,
      )}`,
    );

    const ok = check(res, {
      'users/getReview 200 or 404': (r) => [200, 404].includes(r.status),
    });
    
  }

  
  {
    const res = http.get(`${BASE_URL}/users/stats`);
    const ok = check(res, {
      'users/stats before 200': (r) => r.status === 200,
    });
    
  }

  
  {
    const res = http.get(`${BASE_URL}/pullRequest/stats`);
    const ok = check(res, {
      'pullRequest/stats before 200': (r) => r.status === 200,
    });
    
  }

  
  let reassignedCount = 0;
  {
    const payload = JSON.stringify({
      from_team_name: TEAM_NAME,
      to_team_name: NEW_TEAM_NAME,
    });

    const res = http.post(`${BASE_URL}/team/deactivate`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const ok = check(res, {
      'team/deactivate 200': (r) => {
        if (r.status !== 200) return false;
        const body = r.json();
        reassignedCount = body.reassigned_reviewers || 0;
        return typeof reassignedCount === 'number';
      },
    });
    
  }

  
  {
    const ok = check(null, {
      'reassigned_reviewers > 0': () => reassignedCount > 0,
    });
    
  }

  sleep(1);
}
