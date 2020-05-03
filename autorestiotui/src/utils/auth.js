import axios from 'axios';
import Router from 'vue-router';

export function login() {
}

var router = new Router({
    mode: 'history',
});

export function logout() {
    router.go('/');
}

export function requireAuth(to, from, next) {
    if (!isLoggedIn()) {
        next({
            path: '/',
            query: { redirect: to.fullPath }
        });
    } else {
        next();
    }
}

export function isLoggedIn() {
    const idToken = getIdToken();
    return !!idToken && !isTokenExpired(idToken);
}