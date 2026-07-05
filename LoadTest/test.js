import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    duration: '30s',
    vus: 50,
    thresholds: {
        'http_req_duration': ['p(95)<500'],
        'http_req_failed': ['rate<0.01'],
    },
}

export default function() {
    const checkoutUrl = 'http://localhost:8080/checkout'
    const checkoutRequestPayload = JSON.stringify({
        address_1: '123 Main Street',
        address_2: 'Apt 101',
        state: 'CA',
        zip_code: '1',
        payment_token: '1',
        amount: 1,
        currency: 'usd',
        idempotency_key: '1'
    })
    const checkoutResponse = http.post(checkoutUrl, checkoutRequestPayload, { 'Content-Type': 'application/json' });
    check(checkoutResponse, { 'GET status is 200': (r) => r.status === 200 });
    sleep(1);
}