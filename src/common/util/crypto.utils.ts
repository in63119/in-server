import CryptoJS from 'crypto-js';

export const encrypt = (code: string, secret: string) => {
  const encrypted = CryptoJS.AES.encrypt(code, secret).toString();
  return encrypted;
};

export const decrypt = (encrypted: string, secret: string) => {
  const bytes = CryptoJS.AES.decrypt(encrypted, secret);
  const decrypted = bytes.toString(CryptoJS.enc.Utf8);
  return decrypted;
};

export const sha256 = (data: string) => {
  return CryptoJS.SHA256(data).toString();
};
