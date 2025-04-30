// src/components/StudentDashboard.js

import React, { useContext, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../contexts/AuthContext';
import {
  Container,
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Tabs,
  Tab,
  useTheme,
  useMediaQuery,
  Avatar,
  Paper,
} from '@mui/material';
import { styled } from '@mui/system';
import TodaySchedule from './TodaySchedule'; // reuse tutor's schedule component

// -------------------- Brand Colors --------------------
const brandBlue = '#0e1027';
const brandGold = '#b29600';
const lightBackground = '#fafafa';

// -------------------- Styled Components --------------------
const RootContainer = styled(Box)(() => ({
  minHeight: '100vh',
  backgroundColor: lightBackground,
}));

const StyledAppBar = styled(AppBar)(() => ({
  backgroundColor: brandBlue,
}));

const HeroSection = styled(Box)(({ theme }) => ({
  borderRadius: '8px',
  padding: '40px',
  marginTop: '24px',
  marginBottom: '24px',
  color: '#fff',
  background: `linear-gradient(to bottom right, ${brandBlue}, #2a2f45)`,
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
  [theme.breakpoints.down('md')]: {
    padding: '24px',
    marginTop: '16px',
    marginBottom: '24px',
  },
  [theme.breakpoints.down('sm')]: {
    padding: '16px',
    marginTop: '12px',
    marginBottom: '16px',
  },
}));

const ContentWrapper = styled(Box)(({ theme }) => ({
  backgroundColor: '#fff',
  borderRadius: '16px',
  padding: '24px',
  marginBottom: '40px',
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
  [theme.breakpoints.down('md')]: {
    padding: theme.spacing(2),
    marginBottom: '24px',
  },
  [theme.breakpoints.down('sm')]: {
    padding: theme.spacing(1.5),
    marginBottom: '16px',
  },
}));

const SectionContainer = styled(Paper)(({ theme }) => ({
  padding: '24px',
  borderRadius: '16px',
  backgroundColor: '#fff',
  boxShadow: '0 4px 20px rgba(0,0,0,0.05)',
  [theme.breakpoints.down('md')]: {
    padding: theme.spacing(2),
    borderRadius: '12px',
  },
}));

const SectionTitle = styled(Typography)(() => ({
  marginBottom: '16px',
  fontWeight: 600,
  color: brandBlue,
}));

// -------------------- TabPanel Helper --------------------
function TabPanel({ children, value, index }) {
  return (
    <Box role="tabpanel" hidden={value !== index} sx={{ mt: 2 }}>
      {value === index && children}
    </Box>
  );
}

// -------------------- StudentDashboard --------------------
const StudentDashboard = () => {
  const { user, updateToken } = useContext(AuthContext);
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState(0);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const backendUrl = process.env.REACT_APP_API_BASE_URL;

  useEffect(() => {
    // Could fetch student-specific data here
  }, [user]);

  const handleSignOut = () => {
    localStorage.removeItem('authToken');
    updateToken(null);
    navigate('/');
  };

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  return (
    <RootContainer>
      <StyledAppBar position="static" elevation={3}>
        <Toolbar disableGutters sx={{ px: 2, py: 1, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Typography variant={isMobile ? 'h6' : 'h5'} sx={{ fontWeight: 'bold' }}>
              Welcome, {user?.name || 'Student'}!
            </Typography>
            <Avatar
              src={user?.picture || undefined}
              alt={user?.name || 'Student'}
              sx={{
                bgcolor: user?.picture ? 'transparent' : brandGold,
                color: '#fff',
                width: isMobile ? 40 : 56,
                height: isMobile ? 40 : 56,
              }}
            >
              {!user?.picture && user?.name?.charAt(0).toUpperCase()}
            </Avatar>
          </Box>

          <Button
            onClick={handleSignOut}
            variant="contained"
            sx={{
              backgroundColor: brandGold,
              color: '#fff',
              textTransform: 'none',
              fontWeight: 'bold',
              '&:hover': { backgroundColor: '#d4a100' },
            }}
          >
            Sign Out
          </Button>
        </Toolbar>
      </StyledAppBar>

      <Container maxWidth="xl" sx={{ mt: 3 }}>
        <HeroSection>
          <Typography variant={isMobile ? 'h5' : 'h4'} sx={{ fontWeight: 700 }}>
            Your Schedule Today:
          </Typography>
          {user?.id ? (
            <Box sx={{ mt: 2 }}>
              <TodaySchedule studentId={user.id} backendUrl={backendUrl} />
            </Box>
          ) : (
            <Typography variant="body1" sx={{ opacity: 0.8, mt: 1 }}>
              Loading schedule...
            </Typography>
          )}
        </HeroSection>
      </Container>

      <Container maxWidth="xl">
        <ContentWrapper>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap' }}>
            <Tabs
              value={activeTab}
              onChange={handleTabChange}
              textColor="primary"
              indicatorColor="primary"
              variant={isMobile ? 'fullWidth' : 'standard'}
              sx={{ mb: 2 }}
            >
              <Tab label="Appointments" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="Homework" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="Test Results" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="Resources" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
            </Tabs>
          </Box>

          <TabPanel value={activeTab} index={0}>
            <SectionContainer>
              <SectionTitle variant="h6">Appointments</SectionTitle>
              <Typography>Appointments content goes here.</Typography>
            </SectionContainer>
          </TabPanel>

          <TabPanel value={activeTab} index={1}>
            <SectionContainer>
              <SectionTitle variant="h6">Homework</SectionTitle>
              <Typography>Homework content goes here.</Typography>
            </SectionContainer>
          </TabPanel>

          <TabPanel value={activeTab} index={2}>
            <SectionContainer>
              <SectionTitle variant="h6">Test Results</SectionTitle>
              <Typography>Test Results content goes here.</Typography>
            </SectionContainer>
          </TabPanel>

          <TabPanel value={activeTab} index={3}>
            <SectionContainer>
              <SectionTitle variant="h6">Resources</SectionTitle>
              <Typography>Resources content goes here.</Typography>
            </SectionContainer>
          </TabPanel>
        </ContentWrapper>
      </Container>
    </RootContainer>
  );
};

export default StudentDashboard;
