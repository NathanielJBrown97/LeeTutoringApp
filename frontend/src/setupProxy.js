// src/setupProxy.js

const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
  app.use(
    ['/api', '/internal'],
    createProxyMiddleware({
      target: 'http://localhost:8080', // Ensure this matches your backend port
      changeOrigin: true,
      secure: false,
    })
  );
};
