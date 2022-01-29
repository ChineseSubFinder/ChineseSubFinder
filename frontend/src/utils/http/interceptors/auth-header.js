import { userState } from 'src/store/userState';

export default {
  onRequestFullFilled: (req) => {
    const { accessToken } = userState;
    if (accessToken) {
      req.headers.Authorization = `Bearer ${accessToken}`;
    }
    return req;
  },
};
