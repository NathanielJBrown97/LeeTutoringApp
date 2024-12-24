// src/NoScrollWrapper.js

import React from 'react';
import { GlobalStyles } from '@mui/material';

export default function NoScrollWrapper({ children }) {
  return (
    <>
      <GlobalStyles
        styles={{
          'html, body': {
            margin: 0,
            padding: 0,
            width: '100vw',
            height: '100vh',
            overflow: 'hidden', // Lock scrolling
          },
          '#root': {
            width: '100%',
            height: '100%',
          },
        }}
      />
      {children}
    </>
  );
}
