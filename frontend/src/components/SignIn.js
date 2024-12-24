// src/components/SignIn.js

import React, { useEffect } from 'react';
import { API_BASE_URL } from '../config';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Stack,
  Button,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { styled } from '@mui/system';
import Logo from '../assets/logo.png';
import loginImage from '../assets/login.jpg';

// Importing icons from react-icons
import { FaGoogle, FaMicrosoft, FaFacebook, FaApple, FaYahoo } from 'react-icons/fa';

// Brand colors
const brandBlue = '#0e1027';
const brandGold = '#b29600';

// Define platform-specific colors
const platformColors = {
  google: '#DB4437',
  microsoft: '#F25022',
  facebook: '#3B5998',
  yahoo: '#430297',
  apple: '#000000',
};

// =================== STYLED COMPONENTS ===================

/**
 * RootContainer:
 *  - Desktop: horizontal flex (70% image, 30% sign-in).
 *  - Mobile: entire background is the image; sign-in is centered.
 */
const RootContainer = styled(Box)(({ theme }) => ({
  display: 'flex',
  width: '100vw',
  height: '100vh',
  margin: 0,
  padding: 0,
  overflow: 'hidden', // Lock scrolling

  [theme.breakpoints.down('md')]: {
    backgroundImage: `url(${loginImage})`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    backgroundRepeat: 'no-repeat',
    justifyContent: 'center',
    alignItems: 'center',
  },
}));

/**
 * ImageContainer:
 *  - Desktop only (70% width).
 *  - Hidden on mobile (md and below).
 */
const ImageContainer = styled(Box)(({ theme }) => ({
  width: '70%',
  height: '100%',
  position: 'relative',
  backgroundImage: `url(${loginImage})`,
  backgroundSize: 'cover',
  backgroundPosition: 'center',
  backgroundRepeat: 'no-repeat',

  [theme.breakpoints.down('md')]: {
    display: 'none',
  },
}));

/** 
 * A semi-transparent overlay for the desktop image 
 */
const ImageOverlay = styled(Box)(() => ({
  position: 'absolute',
  inset: 0,
  background: `${brandBlue}80`, // 50% opacity
}));

/**
 * SignInContainer:
 *  - Desktop: 30% width, light background.
 *  - Mobile: fills screen w/ transparent background to show the image.
 */
const SignInContainer = styled(Box)(({ theme }) => ({
  width: '30%',
  height: '100%',
  backgroundColor: '#f9f9f9',
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  padding: theme.spacing(2),

  [theme.breakpoints.down('md')]: {
    width: '100%',
    height: 'auto',
    backgroundColor: 'transparent',
    // Extra bottom padding to avoid iOS bottom bar overlap
    padding: theme.spacing(2, 2, 6),
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
  },
}));

const StyledCard = styled(Card)(({ theme }) => ({
  width: '100%',
  maxWidth: '350px',
  margin: 'auto',
  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.15)',
  borderRadius: '16px',
  backgroundColor: '#fff',

  [theme.breakpoints.down('sm')]: {
    maxWidth: '90%',
  },
}));

const LogoContainer = styled(Box)(() => ({
  display: 'flex',
  justifyContent: 'center',
  marginBottom: '1rem',
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
  backgroundColor: brandBlue,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'flex-start',
  transition: 'background-color 0.3s',
  '&:hover': {
    backgroundColor: hovercolor || brandGold,
  },
  [theme.breakpoints.down('sm')]: {
    fontSize: '14px',
    padding: theme.spacing(1),
  },
}));

const IconWrapper = styled(Box)(({ theme }) => ({
  marginRight: theme.spacing(2),
  display: 'flex',
  alignItems: 'center',
}));

// =================== MAIN COMPONENT ===================
const SignIn = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  // Lock scrolling on mount, restore on unmount
  useEffect(() => {
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = 'auto';
    };
  }, []);

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

  // Platform data (with Yahoo icon)
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
      icon: <FaYahoo size={24} />,
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
      {/* Desktop-Only Image Section */}
      {!isMobile && (
        <ImageContainer>
          <ImageOverlay />
        </ImageContainer>
      )}

      {/* Sign-In Section */}
      <SignInContainer>
        <StyledCard>
          <CardContent>
            {/* Logo */}
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
                  border: `3px solid ${brandGold}`,
                }}
              />
            </LogoContainer>

            {/* Fancy heading with gold bars on each side (fixed widths) */}
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                mb: 3,
              }}
            >
              {/* Left gold bar */}
              <Box
                sx={{
                  width: 80,
                  height: 4,
                  backgroundColor: brandGold,
                  mr: 2,
                }}
              />
              {/* Heading text (no wrapping) */}
              <Typography
                variant={isMobile ? 'h5' : 'h4'}
                sx={{
                  color: brandBlue,
                  fontWeight: 'bold',
                  whiteSpace: 'nowrap', // Force single-line
                }}
              >
                Welcome to Agora
              </Typography>
              {/* Right gold bar */}
              <Box
                sx={{
                  width: 80,
                  height: 4,
                  backgroundColor: brandGold,
                  ml: 2,
                }}
              />
            </Box>

            {/* Subtitle */}
            <Typography
              variant="body1"
              align="center"
              sx={{ mb: 3, color: '#555' }}
            >
              Please sign in using one of the following:
            </Typography>

            {/* Buttons */}
            <Stack spacing={2}>
              {platforms.map((platform) => (
                <StyledButton
                  key={platform.name}
                  onClick={platform.handler}
                  hovercolor={platform.hoverColor}
                  startIcon={
                    platform.icon ? (
                      <IconWrapper>{platform.icon}</IconWrapper>
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
