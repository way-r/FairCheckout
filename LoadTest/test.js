import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    stages: [
        { duration: '10s', target: 50 },
        { duration: '20s', target: 100 },
        { duration: '10s', target: 0 },
    ],
}

const ADDRESS_POOL_SIZE = 100;
const checkoutUrl = 'http://localhost:8080/checkout'

export default function() {
    const randomHouseNumber = Math.floor(Math.random() * ADDRESS_POOL_SIZE) + 1;
    const payload = JSON.stringify({
        payment_id: uuidv4(),
        shipping_address: {
            line_1:  `${randomIntBetween(0, ADDRESS_POOL_SIZE)} main street`,
            line_2: "Apt A1",
            city: "Bellevue",
            zip_code: "98005",
        }
    })
    const checkoutResponse = http.post(checkoutUrl, payload, { 'Content-Type': 'application/json' });
    check(checkoutResponse, { 'GET status is 200': (r) => r.status === 200 });
    sleep(1);
}