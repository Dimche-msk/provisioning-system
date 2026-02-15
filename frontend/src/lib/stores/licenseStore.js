import { writable, derived } from 'svelte/store';

export const licenseInfo = writable({
    tier: 'Free',
    customer_id: '',
    issued_to: '',
    valid_from: null,
    expiry: null,
    support_level: 'Community',
    license_key: ''
});

export const isPro = derived(licenseInfo, ($info) => $info.tier === 'Pro' || $info.tier === 'VIP');
export const supportLevel = derived(licenseInfo, ($info) => $info.support_level);

export async function fetchLicenseStatus() {
    try {
        const response = await fetch('/api/system/stats');
        if (response.ok) {
            const data = await response.json();
            if (data.license) {
                licenseInfo.set(data.license);
            }
        }
    } catch (error) {
        console.error('Failed to fetch license status:', error);
    }
}
