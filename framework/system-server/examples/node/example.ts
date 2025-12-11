// npm install @kubernetes/client-node

import { KubeConfig } from '@kubernetes/client-node';
import axios from 'axios';

export function getTokenFromKubeConfig(): string {
  const kc = new KubeConfig();
  kc.loadFromDefault();
  const user = kc.getCurrentUser();
  return user?.token || '';
}

export function createAuthorizedAxios(baseURL: string) {
  const token = getTokenFromKubeConfig();
  return axios.create({
    baseURL,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}