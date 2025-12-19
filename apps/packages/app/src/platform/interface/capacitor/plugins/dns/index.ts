import { registerPlugin } from '@capacitor/core';
import { DNSServicePlugin } from './definitions';

const DNSService = registerPlugin<DNSServicePlugin>('DNSServicePlugin');

export { DNSService };
