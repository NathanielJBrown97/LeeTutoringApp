// src/components/SignIn.js

import React from 'react';
import { API_BASE_URL } from '../config';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Stack,
  Button,
  Container,
} from '@mui/material';
import { styled } from '@mui/system';
import Logo from '../assets/logo.png'; 
import loginImage from '../assets/login.jpg';

// Importing icons from react-icons
import { FaGoogle, FaMicrosoft, FaFacebook, FaApple } from 'react-icons/fa';

// Define custom colors
const navy = '#001F54';
const cream = '#FFF8E1';
const gold = '#FFD700';

// Define platform-specific colors
const platformColors = {
  google: '#DB4437',
  microsoft: '#F25022',
  facebook: '#3B5998',
  yahoo: '#430297',
  apple: '#000000',
};

// Styled components
const RootContainer = styled(Box)({
  display: 'flex',
  height: '100vh',
  width: '100vw',
  overflow: 'hidden',
});

const ImageContainer = styled(Box)({
  flex: 1,
  backgroundImage: `url(${loginImage})`,
  backgroundSize: 'cover',
  backgroundPosition: 'center',
  backgroundRepeat: 'no-repeat',
  position: 'relative',
});

// Optional: Add a subtle overlay on the image if you want text or branding visible
const ImageOverlay = styled(Box)({
  position: 'absolute',
  inset: 0,
  background: 'rgba(0,0,0,0.3)', // dark semi-transparent overlay
});

const SignInContainer = styled(Box)(({ theme }) => ({
  flex: 1,
  backgroundColor: '#f9f9f9',
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  padding: theme.spacing(4),
}));

const StyledCard = styled(Card)(({ theme }) => ({
  width: '100%',
  maxWidth: '400px',
  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.1)',
  borderRadius: '16px',
  backgroundColor: cream,
}));

const LogoContainer = styled(Box)(({ theme }) => ({
  display: 'flex',
  justifyContent: 'center',
  marginBottom: theme.spacing(3),
}));

const StyledButton = styled(Button, {
  shouldForwardProp: (prop) => prop !== 'hovercolor',
})(({ theme, hovercolor }) => ({
  textTransform: 'none',
  padding: theme.spacing(1.5),
  borderRadius: '8px',
  fontSize: '16px',
  fontWeight: 'bold',
  color: '#fff',
  backgroundColor: navy,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'flex-start',
  transition: 'background-color 0.3s',
  '&:hover': {
    backgroundColor: hovercolor,
  },
}));

const IconWrapper = styled(Box)(({ theme }) => ({
  marginRight: theme.spacing(2),
  display: 'flex',
  alignItems: 'center',
}));

const SignIn = () => {
  // Handler functions
  const handleGoogleLogin = () => {
    window.location.href = `${API_BASE_URL}/internal/googleauth/oauth`;
  };

  const handleMicrosoftLogin = () => {
    window.location.href = `${API_BASE_URL}/internal/microsoftauth/oauth`;
  };

  const handleYahooLogin = () => {
    window.location.href = `${API_BASE_URL}/internal/yahooauth/oauth`;
  };

  const handleFacebookLogin = () => {
    window.location.href = `${API_BASE_URL}/internal/facebookauth/oauth`;
  };

  const handleAppleLogin = () => {
    window.location.href = `${API_BASE_URL}/internal/appleauth/oauth`;
  };

  // Platform data (Yahoo without icon)
  const platforms = [
    {
      name: 'Google',
      handler: handleGoogleLogin,
      icon: <FaGoogle size={24} />,
      hoverColor: platformColors.google,
    },
    {
      name: 'Microsoft',
      handler: handleMicrosoftLogin,
      icon: <FaMicrosoft size={24} />,
      hoverColor: platformColors.microsoft,
    },
    {
      name: 'Yahoo',
      handler: handleYahooLogin,
      icon: null, // No icon for Yahoo for now
      hoverColor: platformColors.yahoo,
    },
    {
      name: 'Facebook',
      handler: handleFacebookLogin,
      icon: <FaFacebook size={24} />,
      hoverColor: platformColors.facebook,
    },
    {
      name: 'Apple',
      handler: handleAppleLogin,
      icon: <FaApple size={24} />,
      hoverColor: platformColors.apple,
    },
  ];

  return (
    <RootContainer>
      {/* Image Section */}
      <ImageContainer>
        <ImageOverlay />
        {/* If you want to add a tagline or logo over the image, you can do so here:
        <Box position="absolute" bottom={40} left={40} color="#fff">
          <Typography variant="h3">Empower your learning</Typography>
        </Box> */}
      </ImageContainer>

      {/* Sign-In Section */}
      <SignInContainer>
        <StyledCard>
          <CardContent>
            <LogoContainer>
              <Box
                component="img"
                src={Logo}
                alt="Agora Logo"
                sx={{
                  width: 80,
                  height: 80,
                  borderRadius: '50%',
                  objectFit: 'cover',
                  border: `4px solid ${gold}`,
                }}
              />
            </LogoContainer>
            <Typography
              variant="h4"
              component="h1"
              align="center"
              gutterBottom
              sx={{ color: navy, fontWeight: 'bold', marginBottom: 2 }}
            >
              Welcome to Agora
            </Typography>
            <Typography
              variant="body1"
              align="center"
              sx={{ mb: 4, color: '#333' }}
            >
              Please sign in using one of the following:
            </Typography>
            <Stack spacing={2}>
              {platforms.map((platform) => (
                <StyledButton
                  key={platform.name}
                  onClick={platform.handler}
                  hovercolor={platform.hoverColor}
                  startIcon={
                    platform.icon ? (
                      <IconWrapper>
                        {platform.icon}
                      </IconWrapper>
                    ) : null
                  }
                  fullWidth
                  aria-label={`Sign in with ${platform.name}`}
                >
                  Sign in with {platform.name}
                </StyledButton>
              ))}
            </Stack>
          </CardContent>
        </StyledCard>
      </SignInContainer>
    </RootContainer>
  );
};

export default SignIn;
