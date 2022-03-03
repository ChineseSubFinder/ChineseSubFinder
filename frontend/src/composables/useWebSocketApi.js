import EventEmitter from 'events';
import { LocalStorage } from 'quasar';
import { SystemMessage } from 'src/utils/Message';
import { onBeforeUnmount, onMounted } from 'vue';

class WSManager extends EventEmitter {
  ws = null;

  url = null;

  autoConnect = true;

  connected = false;

  constructor(url) {
    super();
    this.url = url;
  }

  send(type, data) {
    this.ws.send(
      JSON.stringify({
        type,
        data: JSON.stringify(data),
      })
    );
  }

  connect() {
    const ws = new WebSocket(this.url);
    ws.onopen = () => {
      // 连接成功后自动发送验证信息
      this.send('auth', {
        token: LocalStorage.getItem('token'),
      });
    };

    ws.onmessage = (e) => {
      try {
        const { type, data } = JSON.parse(e.data);
        // console.log(type, data);
        this.emit(type, JSON.parse(data));
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error(error);
      }
    };

    ws.onclose = (e) => {
      this.connected = false;
      if (this.autoConnect) {
        // eslint-disable-next-line no-console
        console.log('Socket is closed. Reconnect will be attempted in 2 second.', e.reason);
        setTimeout(() => {
          this.connect();
        }, 2000);
      }
    };

    ws.onerror = (err) => {
      // eslint-disable-next-line no-console
      console.error('Socket encountered error: ', err.message, 'Closing socket');
      ws.close();
    };

    this.ws = ws;

    this.connected = true;
  }

  // 强制关闭连接
  close() {
    this.autoConnect = false;
    this.ws?.close();
  }
}

// 根据BACKEND_URL配置计算ws地址
export const getWsBaseUrl = () => {
  let result = '';
  const backendUrl = process.env.BACKEND_URL;
  if (!backendUrl) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const { host } = window.location;
    result = `${protocol}//${host}`;
  } else if (backendUrl.startsWith('http')) {
    const protocol = backendUrl.startsWith('https') ? 'wss:' : 'ws:';
    result = `${protocol}//${backendUrl.split('//')[1]}`;
  } else if (backendUrl.startsWith('/')) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    result = `${protocol}//${window.location.host}${backendUrl}`;
  }

  return 'ws://localhost:8080' || result;
};

export const wsManager = new WSManager(`${getWsBaseUrl()}/ws`);

// 处理认证信息
wsManager.on('common_reply', (data) => {
  if (data.message === 'auth error') {
    SystemMessage.error('Websocket验证失败，请重新登录');
  }
});

export const useWebSocketApi = (eventType, eventHandler) => {
  if (wsManager.connected === false) {
    wsManager.connect();
  }
  onMounted(() => {
    wsManager.on(eventType, eventHandler);
  });

  onBeforeUnmount(() => {
    wsManager.off(eventType, eventHandler);
  });
};
