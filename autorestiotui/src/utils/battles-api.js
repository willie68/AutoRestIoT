
import axios from 'axios';

const BASE_URL = 'https://localhost:9443';

export { getPublicStartupBattles, getPrivateStartupBattles };

function getPublicStartupBattles() {
    const url = `${BASE_URL}/api/v1/public/info`;
    return axios.get(url).then(response => response.data);
}

function getPrivateStartupBattles() {
    const url = `${BASE_URL}/api/v1/private/info`;
    return axios.get(url).then(response => response.data);
}