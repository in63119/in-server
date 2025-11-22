// Root
const authRoot = 'auth';

export const routes = {
  auth: {
    root: authRoot,
    signin: `sign-in`,
    signup: `sign-up`,
    options: `options`,
    verify: `verify`,
    registerOptions: `registration/options`,
    registerVerify: `registration/verify`,
    kakao: {
      root: `${authRoot}/kakao`,
      token: `token`,
    },
    github: {
      root: `${authRoot}/github`,
      token: `token`,
    },
  },
};
