import React, { useState, useContext, useEffect } from 'react';
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
  Paper,
  CircularProgress,
  Avatar,
  useTheme,
  useMediaQuery,
  Grid,
  Card,
  CardContent,
  Pagination
} from '@mui/material';
import { styled } from '@mui/system';

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

const SectionTitle = styled(Typography)(({ theme }) => ({
  marginBottom: '16px',
  fontWeight: 600,
  color: brandBlue,
  [theme.breakpoints.down('sm')]: {
    fontSize: '1rem',
    marginBottom: '12px',
  },
}));

// Simple TabPanel helper
function TabPanel(props) {
  const { children, value, index, ...other } = props;
  return (
    <Box role="tabpanel" hidden={value !== index} {...other} sx={{ marginTop: '16px' }}>
      {value === index && children}
    </Box>
  );
}

// -------------------- StudentsTab Component --------------------
// This component fetches the list of associated student IDs and then their detailed data.
// Note that we now destructure tutorEmail from props.
function StudentsTab({ tutorId, tutorEmail, backendUrl }) {
  const [studentIds, setStudentIds] = useState([]);
  const [studentDetails, setStudentDetails] = useState([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const studentsPerPage = 10;

  // Fetch associated students list
  useEffect(() => {
    async function fetchAssociatedStudents() {
      try {
        const token = localStorage.getItem('authToken');
        console.log("StudentsTab: Fetching associated students for tutor", tutorId);
        const res = await fetch(
          `${backendUrl}/api/tutor/fetch-associated-students?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
          {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
          }
        );
  
        if (!res.ok) {
          console.error('StudentsTab: Failed to fetch associated students. Status:', res.status);
          return;
        }
        const ids = await res.json(); // Expected: [{ id: string, name?: string }, ...]
        console.log("StudentsTab: Fetched associated student IDs:", ids);
        setStudentIds(ids);
      } catch (error) {
        console.error('StudentsTab: Error fetching associated students:', error);
      }
    }
    if (tutorId) {
      fetchAssociatedStudents();
    }
  }, [tutorId, tutorEmail, backendUrl]);
  
  // For each associated student, fetch detailed info.
  useEffect(() => {
    async function fetchStudentDetails() {
      try {
        const token = localStorage.getItem('authToken');
        console.log("StudentsTab: Fetching student details for IDs:", studentIds);
        const details = await Promise.all(
          studentIds.map(async (student) => {
            const res = await fetch(
              `${backendUrl}/api/tutor/students/${student.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
              {
                method: 'GET',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`,
                },
              }
            );
            if (!res.ok) {
              console.error(`Failed to fetch details for student ${student.id}. Status: ${res.status}`);
              return null;
            }
            return await res.json();
          })
        );
        setStudentDetails(details.filter(Boolean));
      } catch (error) {
        console.error('StudentsTab: Error fetching student details:', error);
      } finally {
        setLoading(false);
      }
    }
    if (studentIds.length > 0) {
      fetchStudentDetails();
    } else {
      setLoading(false);
    }
  }, [studentIds, tutorId, tutorEmail, backendUrl]);

  // Pagination calculations
  const indexOfLast = currentPage * studentsPerPage;
  const indexOfFirst = indexOfLast - studentsPerPage;
  const currentStudents = studentDetails.slice(indexOfFirst, indexOfLast);
  const totalPages = Math.ceil(studentDetails.length / studentsPerPage);

  const handlePageChange = (event, value) => {
    setCurrentPage(value);
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" sx={{ padding: '24px' }}>
        <CircularProgress />
      </Box>
    );
  }

  if (studentDetails.length === 0) {
    return <Typography variant="body1">No associated students found.</Typography>;
  }

  return (
    <Box>
      <Grid container spacing={2}>
        {currentStudents.map((student) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={student.id}>
            <Card variant="outlined" sx={{ minHeight: '120px', display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', padding: 1 }}>
              <CardContent>
                <Typography variant="h6" align="center">
                  {student.personal?.name || 'Unknown'}
                </Typography>
                <Typography variant="body2" align="center" color="textSecondary">
                  {student.business?.test_focus || 'No Focus'}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
      {totalPages > 1 && (
        <Box display="flex" justifyContent="center" marginTop={2}>
          <Pagination count={totalPages} page={currentPage} onChange={handlePageChange} color="primary" />
        </Box>
      )}
    </Box>
  );
}

// -------------------- TutorDashboard Component --------------------
const TutorDashboard = () => {
  const navigate = useNavigate();
  const { user, updateToken } = useContext(AuthContext);
  const [loading, setLoading] = useState(false);
  const [associationComplete, setAssociationComplete] = useState(false);
  const [activeTab, setActiveTab] = useState(0);
  const [profile, setProfile] = useState(null);

  // Backend URL 
  const backendUrl = process.env.REACT_APP_API_BASE_URL;

  // Responsive layout
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  // Fetch tutor profile on mount.
  useEffect(() => {
    async function fetchTutorProfile() {
      if (!user || !user.id) {
        console.error('TutorDashboard: User info missing from AuthContext.');
        return;
      }
      try {
        const token = localStorage.getItem('authToken');
        console.log("TutorDashboard: Fetching tutor profile for user", user.id);
        const response = await fetch(
          `${backendUrl}/api/tutor/profile?tutorUserID=${encodeURIComponent(user.id)}`,
          {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
          }
        );
        if (!response.ok) {
          console.error('TutorDashboard: Failed to fetch tutor profile. Status:', response.status);
          return;
        }
        const data = await response.json();
        setProfile(data);
        console.log('TutorDashboard: Fetched tutor profile:', data);
      } catch (error) {
        console.error('TutorDashboard: Error fetching tutor profile:', error);
      }
    }
    fetchTutorProfile();
  }, [user, backendUrl]);

  // Trigger association process once profile is loaded.
  useEffect(() => {
    async function associateStudents() {
      if (!profile || !profile.user_id || !profile.email) {
        console.error('TutorDashboard: Tutor profile is not fully loaded.');
        return;
      }
      try {
        const token = localStorage.getItem('authToken');
        console.log("TutorDashboard: Associating students for tutor", profile.user_id);
        const response = await fetch(
          `${backendUrl}/api/tutor/associate-students?tutorUserID=${encodeURIComponent(profile.user_id)}&tutorEmail=${encodeURIComponent(profile.email)}`,
          {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
          }
        );
        if (!response.ok) {
          console.error('TutorDashboard: Failed to associate students. Status:', response.status);
        } else {
          console.log('TutorDashboard: Successfully associated students.');
        }
      } catch (error) {
        console.error('TutorDashboard: Error while associating students:', error);
      } finally {
        setAssociationComplete(true);
      }
    }
    if (profile) {
      associateStudents();
    }
  }, [profile, backendUrl]);

  const handleSignOut = () => {
    localStorage.removeItem('authToken');
    updateToken(null);
    navigate('/');
  };

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh" sx={{ backgroundColor: lightBackground }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <RootContainer>
      {/* ---------- AppBar with Personalized Welcome ---------- */}
      <StyledAppBar position="static" elevation={3}>
        <Toolbar disableGutters sx={{ px: 2, py: 1 }}>
          {isMobile ? (
            <Box sx={{ width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%', mb: 1 }}>
                <Typography variant="h6" sx={{ fontWeight: 'bold' }}>
                  Welcome, {profile ? profile.name : 'Tutor'}!
                </Typography>
                <Button
                  onClick={handleSignOut}
                  variant="contained"
                  sx={{
                    backgroundColor: brandGold,
                    color: '#fff',
                    fontWeight: 'bold',
                    textTransform: 'none',
                    '&:hover': { backgroundColor: '#d4a100' },
                  }}
                >
                  Sign Out
                </Button>
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', py: 1 }}>
                <Avatar
                  src={profile && profile.picture ? profile.picture : undefined}
                  alt={profile ? profile.name : 'Tutor'}
                  sx={{
                    bgcolor: profile && profile.picture ? 'transparent' : brandGold,
                    color: '#fff',
                    width: 48,
                    height: 48,
                  }}
                >
                  {profile && !profile.picture && profile.name ? profile.name.charAt(0).toUpperCase() : 'T'}
                </Avatar>
              </Box>
            </Box>
          ) : (
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%', px: 2, py: 1 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Typography variant="h5" sx={{ fontWeight: 'bold' }}>
                  Welcome, {profile ? profile.name : 'Tutor'}!
                </Typography>
                <Avatar
                  src={profile && profile.picture ? profile.picture : undefined}
                  alt={profile ? profile.name : 'Tutor'}
                  sx={{
                    bgcolor: profile && profile.picture ? 'transparent' : brandGold,
                    color: '#fff',
                    width: 56,
                    height: 56,
                  }}
                >
                  {profile && !profile.picture && profile.name ? profile.name.charAt(0).toUpperCase() : 'T'}
                </Avatar>
              </Box>
              <Button
                onClick={handleSignOut}
                variant="contained"
                sx={{
                  backgroundColor: brandGold,
                  color: '#fff',
                  fontWeight: 'bold',
                  textTransform: 'none',
                  '&:hover': { backgroundColor: '#d4a100' },
                }}
              >
                Sign Out
              </Button>
            </Box>
          )}
        </Toolbar>
      </StyledAppBar>

      {/* ---------- Hero Section ---------- */}
      <Container maxWidth="xl" sx={{ marginTop: '24px' }}>
        <HeroSection>
          <Typography variant="h4" sx={{ fontWeight: 700 }}>
            Your Schedule Today:
          </Typography>
          <Typography variant="body1" sx={{ opacity: 0.9, marginTop: '8px' }}>
            Here There Will Be a Visual Graphic Showing Upcoming Appointments...
          </Typography>
        </HeroSection>
      </Container>

      {/* ---------- Main Content (Tabs etc.) ---------- */}
      <Container maxWidth="xl">
        <ContentWrapper>
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Tabs
              value={activeTab}
              onChange={handleTabChange}
              textColor="primary"
              indicatorColor="primary"
              variant="scrollable"
              scrollButtons="auto"
              sx={{ marginBottom: '16px' }}
            >
              <Tab label="My Students" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="My Schedule" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="Tutor Tools" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
            </Tabs>
          </Box>

          <TabPanel value={activeTab} index={0}>
            <SectionContainer>
              <SectionTitle variant="h6">My Students</SectionTitle>
              {profile && profile.user_id && profile.email ? (
                associationComplete ? (
                  <StudentsTab 
                    tutorId={profile.user_id} 
                    tutorEmail={profile.email} 
                    backendUrl={backendUrl} 
                  />
                ) : (
                  <Typography variant="body1">Updating your student associations...</Typography>
                )
              ) : (
                <Typography variant="body1">Loading your students...</Typography>
              )}
            </SectionContainer>
          </TabPanel>

          <TabPanel value={activeTab} index={1}>
            <SectionContainer>
              <SectionTitle variant="h6">Today's Schedule</SectionTitle>
              <Typography variant="body1">
                No appointments scheduled for today.
              </Typography>
            </SectionContainer>
          </TabPanel>

          <TabPanel value={activeTab} index={2}>
            <SectionContainer>
              <SectionTitle variant="h6">Tutor Tools</SectionTitle>
              <Typography variant="body1">
                Tutor tools coming soon.
              </Typography>
            </SectionContainer>
          </TabPanel>
        </ContentWrapper>
      </Container>
    </RootContainer>
  );
};

export default TutorDashboard;
