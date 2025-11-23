import { register, init, getLocaleFromNavigator } from 'svelte-i18n';
import { browser } from '$app/environment';

register('en', () => import('./locales/en.json'));
register('ru', () => import('./locales/ru.json'));

init({
    fallbackLocale: 'ru',
    initialLocale: browser ? (localStorage.getItem('locale') || 'ru') : 'ru',
});
