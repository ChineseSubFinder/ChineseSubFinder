/**
 * 消息提醒
 */
import { Notify } from 'quasar';

const MESSAGE_DURATION = 500;

export const SystemMessage = {
  success(message, options = {}) {
    Notify.create({
      message,
      icon: 'check_circle',
      position: 'top',
      color: 'positive',
      timeout: MESSAGE_DURATION,
      ...options,
    });
  },

  warn(message, options = {}) {
    Notify.create({
      message,
      icon: 'warning',
      position: 'top',
      color: 'warning',
      timeout: MESSAGE_DURATION,
      ...options,
    });
  },

  info(message, options = {}) {
    Notify.create({
      message,
      icon: 'info',
      position: 'top',
      color: 'info',
      timeout: MESSAGE_DURATION,
      ...options,
    });
  },

  error(message, options = {}) {
    Notify.create({
      message,
      icon: 'error',
      position: 'top',
      color: 'negative',
      timeout: 2000,
      ...options,
    });
  },
};
