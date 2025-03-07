import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { AuthContext } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import ben from '../assets/ben.jpg';
import edward from '../assets/edward.jpg';
import kieran from '../assets/kieran.jpg';
import kyra from '../assets/kyra.jpg';
import omar from '../assets/omar.jpg';
import patrick from '../assets/patrick.jpg';
import eli from '../assets/eli.jpg';

// --- Recharts ---
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';

import {
  Container,
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Grid,
  Select,
  MenuItem,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
  Divider,
  Tabs,
  Tab,
  Avatar,
  ListItemText,
  IconButton,
  Slide,
  useMediaQuery,
  useTheme,
  Chip,
  TextField,
} from '@mui/material';
import { styled } from '@mui/system';
import { Collapse } from '@mui/material';

// MUI Icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';

// Charts
import SuperScoreChart from './SuperScoreChart';

// -------------------- Brand Colors --------------------
const brandBlue = '#0e1027';
const brandGold = '#b29600';
const lightBackground = '#fafafa';
const brandGoldLight = '#d4a100';
const brandGoldLighter = '#f5dd5c';
const brandBlueLight = '#2a2f45';

const tutorImages = {
  ben,
  edward,
  kieran,
  kyra,
  omar,
  patrick,
  eli,
};

// The tab labels in an array (for desktop tabs or mobile dropdown).
const tabLabels = [
  'Recent Appointments',
  'School Goals',
  'Student Profile',
  'Test Data',
];

// -------------------- Styled Components --------------------
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

const AppointmentCard = styled(Paper)(({ theme }) => ({
  borderRadius: '12px',
  padding: '16px',
  backgroundColor: '#fff',
  boxShadow: '0 2px 12px rgba(0,0,0,0.1)',
  marginBottom: '16px',
  [theme.breakpoints.down('sm')]: {
    padding: theme.spacing(1.5),
    marginBottom: theme.spacing(1),
  },
}));

const RootContainer = styled(Box)(() => ({
  minHeight: '100vh',
  backgroundColor: lightBackground,
}));

function TabPanel(props) {
  const { children, value, index, ...other } = props;
  return (
    <Box
      role="tabpanel"
      hidden={value !== index}
      {...other}
      sx={{ marginTop: '16px' }}
    >
      {value === index && children}
    </Box>
  );
}

// -- Helper function to safely parse integers --
const parseScore = (scoreStr) => {
  // Attempt to parse a number from the string. Returns NaN if invalid.
  const val = parseInt(scoreStr, 10);
  return isNaN(val) ? null : val;
};

// Percentage change helper
function computePercentageChange(oldVal, newVal) {
  // If oldVal is null, zero, or invalid, we can't do a % difference
  if (!oldVal || oldVal <= 0) return null;
  const diff = newVal - oldVal;
  const percent = (diff / oldVal) * 100; // e.g. +25 means +25%
  return percent;
}

// A combined function that returns a Chip showing both the numeric difference
// (+2 or -1) AND the % difference (+10.0% or -5.0%).
function renderChangeChip(oldVal, newVal) {
  if (oldVal === null || newVal === null) return '—'; // baseline or invalid

  const diff = newVal - oldVal; // e.g. +2 or -3
  // We won't show the percentage anymore; just the arrow + points.
  const arrowSymbol = diff > 0 ? '▲' : '▼';
  const signDiff = diff > 0 ? `+${diff}` : diff.toString(); // e.g. +2 or -2

  // Choose a color: positive => brandBlue; negative => brandGold.
  const chipBgColor = diff >= 0 ? '#18a558' : brandGold;
  // White text so it's visible on the colored background.
  const combinedLabel = `${arrowSymbol} ${signDiff}`;

  return (
    <Chip
      label={combinedLabel}
      size="small"
      sx={{
        fontWeight: 600,
        backgroundColor: chipBgColor,
        color: '#fff',
      }}
    />
  );
}

const ParentDashboard = () => {
  const authState = useContext(AuthContext);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm')); // phone


  const [attemptedAutoAssociation, setAttemptedAutoAssociation] = useState(false);
  const [associatedStudents, setAssociatedStudents] = useState([]);
  const [selectedStudentID, setSelectedStudentID] = useState(null);
  const [studentData, setStudentData] = useState(null);
  const testFocus = (studentData?.business?.test_focus || 'TBD').toUpperCase();
  useEffect(() => {
    console.log('testFocus is: ', testFocus);
  }, [testFocus]);

  const [parentName, setParentName] = useState('Parent');
  const [parentPicture, setParentPicture] = useState(null);
  const [activeTab, setActiveTab] = useState(0);
  const [loading, setLoading] = useState(true);

  // Parent Profile Editing
  const [editMode, setEditMode] = useState(false);
  const [invoiceEmailValue, setInvoiceEmailValue] = useState('');
  const [emailError, setEmailError] = useState('');
  const [confirmationMode, setConfirmationMode] = useState(false);

  // For the 3-appointment scroller
  const [startIndex, setStartIndex] = useState(0);
  const [scrollDirection, setScrollDirection] = useState('down');

  const navigate = useNavigate();

  const [parentID, setParentID] = useState("");
  const [parentInvoiceEmail, setParentInvoiceEmail] = useState("");


  // -------------- Data Fetching --------------

  // Fetch parent data on mount
  useEffect(() => {
    const token = localStorage.getItem("authToken");
    if (!token) {
      navigate("/");
      return;
    }

    fetch(`${API_BASE_URL}/api/parent`, {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch parent");
        return res.json();
      })
      .then((data) => {
        setParentID(data.user_id); // Save the parent's document ID
        setParentName(data.name || "Parent");
        setParentPicture(data.picture || null);
        setParentInvoiceEmail(data.invoice_email || ""); // Store the invoice email
      })
      .catch((err) => {
        console.error("Error fetching parent data:", err);
      });
  }, [navigate]);


    // 2) Attempt Automatic Association (only once)
    useEffect(() => {
      const token = localStorage.getItem('authToken');
      if (!token || attemptedAutoAssociation) return;

      fetch(`${API_BASE_URL}/api/attemptAutomaticAssociation`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      })
        .then(async (res) => {
          if (!res.ok) {
            throw new Error('Auto-association attempt failed');
          }
          return res.json();
        })
        .then((data) => {
          console.log('Auto-association response:', data);
          // E.g. some data structure that shows success or not
          setAttemptedAutoAssociation(true);
          // Now fetch the associated students
          fetchAssociatedStudents(token);
        })
        .catch((err) => {
          // If the attempt fails, navigate to StudentIntake
          console.error('Auto-association error:', err);
          navigate('/studentintake');
        });
    }, [attemptedAutoAssociation, navigate]);

  // 3) Fetch associated students
  const fetchAssociatedStudents = (token) => {
    fetch(`${API_BASE_URL}/api/associated-students`, {
      method: 'GET',
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => res.json())
      .then((data) => {
        if (!data.associatedStudents || data.associatedStudents.length === 0) {
          // If no associated students, redirect to StudentIntake
          console.warn('No students found, redirecting...');
          navigate('/studentintake');
          return;
        }
        setAssociatedStudents(data.associatedStudents);
        setSelectedStudentID(data.associatedStudents[0]);
      })
      .catch((err) => {
        console.error('fetchAssociatedStudents error:', err);
        // Optionally also navigate to StudentIntake if error
        navigate('/studentintake');
      });
  };

  // 4) Whenever the selectedStudentID changes, fetch their data
  useEffect(() => {
    if (!selectedStudentID) return;
    const token = localStorage.getItem('authToken');

    setLoading(true);
    fetch(`${API_BASE_URL}/api/students/${selectedStudentID}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => res.json())
      .then((data) => {
        setStudentData(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Error fetching student data:', err);
        setLoading(false);
      });
  }, [selectedStudentID]);

  // -------------- Handlers --------------

  // For parent profile editing

  // Email validation helper
  const isValidEmail = (email) => {
    // Simple regex for basic email validation
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  };

  // Handler for clicking the "Edit" button
  const handleEditClick = () => {
    setEditMode(true);
    setInvoiceEmailValue(studentData?.parent?.invoice_email || '');
    setEmailError('');
    setConfirmationMode(false);
  };

  // Update input value as user types
  const handleInputChange = (e) => {
    setInvoiceEmailValue(e.target.value);
  };

  // Handler for "Change" button click
  const handleSaveClick = () => {
    if (!isValidEmail(invoiceEmailValue)) {
      setEmailError('Please enter a valid email address.');
      return;
    }
    setEmailError('');
    // Instead of immediately saving, show the confirmation check.
    setConfirmationMode(true);
  };

  // Handler for when the user confirms the change
  const handleConfirmSave = () => {
    const token = localStorage.getItem("authToken");
    if (!parentID) {
      console.error("Parent ID is missing");
      return;
    }
    fetch(`${API_BASE_URL}/api/updateInvoiceEmail`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        parent_id: parentID, // Use the parent's ID we stored
        invoice_email: invoiceEmailValue,
      }),
    })
      .then((res) => {
        console.log("Response status:", res.status);
        return res.json();
      })
      .then((data) => {
        console.log("Invoice email updated:", data);
        // Optionally update your studentData if it contains a parent subdoc
        setStudentData((prev) => ({
          ...prev,
          parent: {
            ...prev.parent,
            invoice_email: invoiceEmailValue,
          },
        }));
        setParentInvoiceEmail(invoiceEmailValue);
        setConfirmationMode(false);
        setEditMode(false);
      })
      .catch((error) => {
        console.error("Error updating invoice email:", error);
      });
  };
  
  

  // Handler for the "Go Back" button in the confirmation view.
  const handleCancelConfirmation = () => {
    // Return to the edit view so the user can make corrections.
    setConfirmationMode(false);
  };

    if (loading || !studentData) {
      return (
        <Box
          display="flex"
          justifyContent="center"
          alignItems="center"
          height="100vh"
          sx={{ backgroundColor: lightBackground }}
        >
          <CircularProgress />
        </Box>
      );
    }


  const handleSignOut = () => {
    localStorage.removeItem('authToken');
    authState.updateToken(null);
    navigate('/');
  };

  const handleStudentChange = (event) => {
    const newStudentID = event.target.value;
    setSelectedStudentID(newStudentID);
    navigate(`/parentdashboard?studentID=${newStudentID}`);
  };

  // For desktop tabs
  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  // Book With A Tutor
  const handleNavigateToBooking = () => {
    if (selectedStudentID) {
      navigate(`/booking?studentID=${selectedStudentID}`);
    } else {
      navigate('/booking');
    }
  };

  // 3-appointment scroller
  const handlePrevAppointment = () => {
    if (startIndex > 0) {
      setScrollDirection('up');
      setStartIndex(startIndex - 1);
    }
  };
  const handleNextAppointment = (appointmentsCount) => {
    if (startIndex + 3 < appointmentsCount) {
      setScrollDirection('down');
      setStartIndex(startIndex + 1);
    }
  };

  // -------------- Loading / Early Return --------------
  if (loading || !studentData) {
    return (
      <Box
        display="flex"
        justifyContent="center"
        alignItems="center"
        height="100vh"
        sx={{ backgroundColor: lightBackground }}
      >
        <CircularProgress />
      </Box>
    );
  }

  // -------------- After we have studentData --------------
  // Test data separation
  const testData = studentData.testData || [];
  const satTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('SAT') || upperTest.includes('PSAT');
  });
  const actTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('ACT');
  });

  // Ascending arrays for line charts
  const ascendingSatTests = [...satTests].sort(
    (a, b) => new Date(a.date || 0) - new Date(b.date || 0)
  );
  const ascendingActTests = [...actTests].sort(
    (a, b) => new Date(a.date || 0) - new Date(b.date || 0)
  );

  // Build line data for SAT (REMOVE Reading & Writing from the chart)
  const lineDataSAT = ascendingSatTests.map((testDoc, idx) => {
    const { EBRW, Math, SAT_Total } = renderSATScores(testDoc);
    return {
      name: testDoc.date || `Test #${idx + 1}`,
      EBRW: parseScore(EBRW) || 0,
      Math: parseScore(Math) || 0,
      Total: parseScore(SAT_Total) || 0,
    };
  });

  // Build line data for ACT
  const lineDataACT = ascendingActTests.map((testDoc, idx) => {
    const { English, MathVal, Reading, Science, ACT_Total } =
      renderACTScores(testDoc);
    return {
      name: testDoc.date || `Test #${idx + 1}`,
      English: parseScore(English) || 0,
      MathVal: parseScore(MathVal) || 0,
      Reading: parseScore(Reading) || 0,
      Science: parseScore(Science) || 0,
      Total: parseScore(ACT_Total) || 0,
    };
  });

  // Sort appointments (descending)
  const sortedAppointments = [...(studentData.homeworkCompletion || [])].sort(
    (a, b) => {
      const dateA = a.date ? new Date(a.date) : new Date(0);
      const dateB = b.date ? new Date(b.date) : new Date(0);
      return dateB - dateA;
    }
  );
  const appointmentsToShow = sortedAppointments.slice(startIndex, startIndex + 3);

  // Filter test dates (only upcoming)
  const testDates = (studentData.testDates || [])
    .filter((test) => {
      const testDateStr = test.test_date;
      if (!testDateStr) return false;
      const testDate = new Date(testDateStr);
      if (isNaN(testDate.getTime())) return false;
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      return testDate >= today;
    })
    .sort((a, b) => new Date(a.test_date) - new Date(b.test_date));

  // Score parsing
  function renderSATScores(testDoc) {
    // We won't display Reading/Writing, but parse them anyway
    let EBRW = '';
    let Math = '';
    let Reading = '';
    let Writing = '';
    let SAT_Total = '';
    if (Array.isArray(testDoc.SAT_Scores) && testDoc.SAT_Scores.length > 0) {
      [EBRW, Math, Reading, Writing, SAT_Total] = testDoc.SAT_Scores;
    } else if (Array.isArray(testDoc.SAT) && testDoc.SAT.length > 0) {
      [EBRW, Math, Reading, Writing, SAT_Total] = testDoc.SAT;
    }
    return { EBRW, Math, Reading, Writing, SAT_Total };
  }

  function renderACTScores(testDoc) {
    // We'll store the second index as MathVal for clarity
    let English = '';
    let MathVal = '';
    let Reading = '';
    let Science = '';
    let ACT_Total = '';
    if (Array.isArray(testDoc.ACT_Scores) && testDoc.ACT_Scores.length > 0) {
      [English, MathVal, Reading, Science, ACT_Total] = testDoc.ACT_Scores;
    } else if (Array.isArray(testDoc.ACT) && testDoc.ACT.length > 0) {
      [English, MathVal, Reading, Science, ACT_Total] = testDoc.ACT;
    }
    return { English, MathVal, Reading, Science, ACT_Total };
  }

  // (Optional) extra helpers for behind-the-scenes computations:
  function getActPercentChanges(currentDoc, prevDoc) {
    if (!prevDoc) {
      return {
        ePercent: null,
        mPercent: null,
        rPercent: null,
        sPercent: null,
        tPercent: null,
      };
    }

    const currScores = renderACTScores(currentDoc);
    const prevScores = renderACTScores(prevDoc);

    const ePercent = computePercentageChange(
      parseScore(prevScores.English),
      parseScore(currScores.English)
    );
    const mPercent = computePercentageChange(
      parseScore(prevScores.MathVal),
      parseScore(currScores.MathVal)
    );
    const rPercent = computePercentageChange(
      parseScore(prevScores.Reading),
      parseScore(currScores.Reading)
    );
    const sPercent = computePercentageChange(
      parseScore(prevScores.Science),
      parseScore(currScores.Science)
    );
    const tPercent = computePercentageChange(
      parseScore(prevScores.ACT_Total),
      parseScore(currScores.ACT_Total)
    );

    return { ePercent, mPercent, rPercent, sPercent, tPercent };
  }

  function getSatPercentChanges(currentDoc, prevDoc) {
    if (!prevDoc) {
      return {
        ePercent: null,
        mPercent: null,
        rPercent: null,
        wPercent: null,
        tPercent: null,
      };
    }

    const currScores = renderSATScores(currentDoc);
    const prevScores = renderSATScores(prevDoc);

    const ePercent = computePercentageChange(
      parseScore(prevScores.EBRW),
      parseScore(currScores.EBRW)
    );
    const mPercent = computePercentageChange(
      parseScore(prevScores.Math),
      parseScore(currScores.Math)
    );
    // Reading/Writing are not displayed, but could be computed
    const rPercent = computePercentageChange(
      parseScore(prevScores.Reading),
      parseScore(currScores.Reading)
    );
    const wPercent = computePercentageChange(
      parseScore(prevScores.Writing),
      parseScore(currScores.Writing)
    );
    const tPercent = computePercentageChange(
      parseScore(prevScores.SAT_Total),
      parseScore(currScores.SAT_Total)
    );

    return { ePercent, mPercent, rPercent, wPercent, tPercent };
  }

  return (
    <RootContainer>
      {/* ---------- AppBar ---------- */}
      <StyledAppBar position="static" elevation={3}>
        <Toolbar
          disableGutters
          sx={{
            px: 2,
            py: 1,
          }}
        >
          {isMobile ? (
            <Box sx={{ width: '100%' }}>
              {/* Top row: Welcome + Sign Out (single line) */}
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  width: '100%',
                  mb: 1,
                  whiteSpace: 'nowrap',
                }}
              >
                <Typography
                  variant="h6"
                  sx={{
                    fontWeight: 'bold',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  Welcome, {parentName}!
                </Typography>

                <Button
                  onClick={handleSignOut}
                  variant="contained"
                  sx={{
                    backgroundColor: brandGold,
                    color: '#fff',
                    fontWeight: 'bold',
                    textTransform: 'none',
                    whiteSpace: 'nowrap',
                    mr: 0.5,
                    '&:hover': {
                      backgroundColor: '#d4a100',
                    },
                  }}
                >
                  Sign Out
                </Button>
              </Box>

              {/* Second row: Avatar only */}
              <Box sx={{ display: 'flex', alignItems: 'center', py: 1 }}>
                <Avatar
                  src={parentPicture || undefined}
                  alt={parentName}
                  sx={{
                    bgcolor: parentPicture ? 'transparent' : brandGold,
                    color: '#fff',
                  }}
                >
                  {!parentPicture && parentName.charAt(0).toUpperCase()}
                </Avatar>
              </Box>
            </Box>
          ) : (
            // Desktop Layout
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'row',
                alignItems: 'center',
                justifyContent: 'space-between',
                width: '100%',
                px: 2,
                py: 1,
              }}
            >
              {/* Left side: Welcome + Avatar */}
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Typography
                  variant="h5"
                  sx={{ fontWeight: 'bold', whiteSpace: 'nowrap' }}
                >
                  Welcome, {parentName}!
                </Typography>
                <Avatar
                  src={parentPicture || undefined}
                  alt={parentName}
                  sx={{
                    bgcolor: parentPicture ? 'transparent' : brandGold,
                    color: '#fff',
                  }}
                >
                  {!parentPicture && parentName.charAt(0).toUpperCase()}
                </Avatar>
              </Box>

              {/* Right side: Sign Out */}
              <Button
                onClick={handleSignOut}
                variant="contained"
                sx={{
                  backgroundColor: brandGold,
                  color: '#fff',
                  fontWeight: 'bold',
                  textTransform: 'none',
                  whiteSpace: 'nowrap',
                  mr: 0.5,
                  '&:hover': {
                    backgroundColor: '#d4a100',
                  },
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
          {/* Single line: "Dashboard Overview for:" + dropdown */}
          <Box
            display="flex"
            alignItems="center"
            flexWrap="nowrap"
            gap={2}
            sx={{ whiteSpace: 'nowrap' }}
          >
            <Typography
              variant={isMobile ? 'h6' : 'h5'}
              sx={{ fontWeight: 700, m: 0 }}
            >
              Overview for:
            </Typography>

            <Select
              size="small"
              value={selectedStudentID}
              onChange={handleStudentChange}
              variant="outlined"
              sx={{
                height: 40,
                backgroundColor: '#fff',
                borderRadius: '8px',
                fontWeight: 500,
                minWidth: isMobile ? 140 : 180,
              }}
            >
              {associatedStudents.map((student) => (
                <MenuItem key={student} value={student}>
                  {student}
                </MenuItem>
              ))}
            </Select>
          </Box>

          {/* Second row: descriptive text */}
          <Box mt={2}>
            <Typography
              variant={isMobile ? 'body1' : 'h6'}
              sx={{ opacity: 0.9 }}
            >
              View your student's recent appointments, goals, and test data.
            </Typography>
          </Box>
        </HeroSection>
      </Container>

      {/* ------------- Main Content ------------- */}
      <Container maxWidth="xl">
        <ContentWrapper>
          {/* Desktop vs. Mobile tabs */}
          {!isMobile ? (
            // Desktop: normal MUI tabs
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
                {tabLabels.map((label, idx) => (
                  <Tab
                    key={label}
                    label={label}
                    sx={{ textTransform: 'none', fontWeight: 'bold' }}
                  />
                ))}
              </Tabs>

              <Button
                onClick={handleNavigateToBooking}
                variant="contained"
                sx={{
                  backgroundColor: brandBlue,
                  color: '#fff',
                  textTransform: 'none',
                  fontWeight: 'bold',
                  '&:hover': {
                    backgroundColor: '#1c2231',
                  },
                }}
              >
                Book With A Tutor
              </Button>
            </Box>
          ) : (
            // Mobile: drop-down for tab navigation
            <Box display="flex" alignItems="center" justifyContent="space-between" gap={2}>
              <Select
                value={String(activeTab)}
                onChange={(e) => {
                  setActiveTab(Number(e.target.value));
                }}
                sx={{ minWidth: 220 }}
              >
                {tabLabels.map((label, idx) => (
                  <MenuItem key={label} value={String(idx)}>
                    {label}
                  </MenuItem>
                ))}
              </Select>

              <Button
                onClick={handleNavigateToBooking}
                variant="contained"
                sx={{
                  backgroundColor: brandBlue,
                  color: '#fff',
                  textTransform: 'none',
                  fontWeight: 'bold',
                  '&:hover': {
                    backgroundColor: '#1c2231',
                  },
                }}
              >
                Book With A Tutor
              </Button>
            </Box>
          )}

          {/* =================== TAB PANELS =================== */}

          <TabPanel value={activeTab} index={0}>
  <SectionContainer>
    <SectionTitle variant="h6">Recent Appointments</SectionTitle>
    <Divider sx={{ marginBottom: '16px' }} />

    <Box display="flex" flexDirection="column" alignItems="center">
      {/* Up Arrow */}
      <IconButton
        onClick={handlePrevAppointment}
        disabled={startIndex === 0}
        sx={{ mb: 2 }}
      >
        <KeyboardArrowUpIcon fontSize="large" />
      </IconButton>

      {/* Desktop: Slide with ±10% offset; Mobile: no animation */}
      {!isMobile ? (
        <Slide
          key={startIndex}
          in
          direction={scrollDirection === 'down' ? 'down' : 'up'}
          timeout={300}
          mountOnEnter
          unmountOnExit
          onEnter={(node) => {
            const offset = '10%';
            node.style.transform =
              scrollDirection === 'down'
                ? `translateY(-${offset})`
                : `translateY(${offset})`;
          }}
          onEntering={(node) => {
            node.style.transform = 'translateY(0%)';
          }}
        >
          <Box sx={{ maxWidth: 400 }}>
            {appointmentsToShow.length > 0 ? (
              appointmentsToShow.map((appt, index) => {
                // 1) Format the date
                const parsedDate = appt.date ? new Date(appt.date) : null;
                const formattedDate = parsedDate
                  ? parsedDate.toLocaleDateString(undefined, {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })
                  : 'N/A';

                // 2) Homework percentage
                const percentage = `${appt.percentage ?? 0}%`;

                // 3) Convert decimal hours to readable format
                const formatDuration = (decimalHours) => {
                  if (!decimalHours || isNaN(decimalHours)) return 'N/A';
                  const totalMinutes = Math.round(decimalHours * 60);
                  const h = Math.floor(totalMinutes / 60);
                  const m = totalMinutes % 60;

                  if (h === 0 && m > 0) {
                    return `${m} Minute${m === 1 ? '' : 's'}`;
                  } else if (h > 0 && m === 0) {
                    return `${h} Hour${h === 1 ? '' : 's'}`;
                  } else if (h > 0 && m > 0) {
                    return `${h} Hour${h === 1 ? '' : 's'} and ${m} Minute${
                      m === 1 ? '' : 's'
                    }`;
                  }
                  return 'N/A';
                };
                const displayDuration = formatDuration(Number(appt.duration));

                // 4) Attendance status
                const status = appt.attendance || 'N/A';

                // 5) Tutor name -> dynamic image
                const tutorName = (appt.tutor || '').toLowerCase();
                const tutorImage =
                  tutorImages[tutorName] ||
                  'https://via.placeholder.com/48?text=Tutor';

                return (
                  <AppointmentCard key={index} sx={{ position: 'relative' }}>
                    <Typography
                      variant="subtitle1"
                      sx={{ fontWeight: 'bold', mb: 1 }}
                    >
                      Appointment Date: {formattedDate}
                    </Typography>
                    <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                      Homework Completed: {percentage}
                    </Typography>
                    <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                      Duration: {displayDuration}
                    </Typography>
                    <Typography variant="body2" sx={{ color: '#333' }}>
                      Status: {status}
                    </Typography>

                    {/* Bottom-right Tutor Image & Name */}
                    <Box
                      sx={{
                        position: 'absolute',
                        // Decrease this to make the avatar appear lower
                        bottom: 8, 
                        right: 16,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                      }}
                    >
                      <Avatar
                        src={tutorImage}
                        alt={appt.tutor || 'Tutor'}
                        sx={{ width: 48, height: 48, mb: 1 }}
                      />
                      <Typography variant="caption" sx={{ fontWeight: 'bold' }}>
                        {appt.tutor || 'Tutor'}
                      </Typography>
                    </Box>
                  </AppointmentCard>
                );
              })
            ) : (
              <Typography variant="body2" color="textSecondary">
                No recent appointments available.
              </Typography>
            )}
          </Box>
        </Slide>
      ) : (
        // ---------- Mobile: No slide animation ----------
        <Box sx={{ maxWidth: '100%' }}>
          {appointmentsToShow.length > 0 ? (
            appointmentsToShow.map((appt, index) => {
              const parsedDate = appt.date ? new Date(appt.date) : null;
              const formattedDate = parsedDate
                ? parsedDate.toLocaleDateString(undefined, {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                  })
                : 'N/A';

              const percentage = `${appt.percentage ?? 0}%`;
              const formatDuration = (decimalHours) => {
                if (!decimalHours || isNaN(decimalHours)) return 'N/A';
                const totalMinutes = Math.round(decimalHours * 60);
                const h = Math.floor(totalMinutes / 60);
                const m = totalMinutes % 60;

                if (h === 0 && m > 0) {
                  return `${m} Minute${m === 1 ? '' : 's'}`;
                } else if (h > 0 && m === 0) {
                  return `${h} Hour${h === 1 ? '' : 's'}`;
                } else if (h > 0 && m > 0) {
                  return `${h} Hour${h === 1 ? '' : 's'} and ${m} Minute${
                    m === 1 ? '' : 's'
                  }`;
                }
                return 'N/A';
              };
              const displayDuration = formatDuration(Number(appt.duration));
              const status = appt.attendance || 'N/A';

              const tutorName = (appt.tutor || '').toLowerCase();
              const tutorImage =
                tutorImages[tutorName] ||
                'https://via.placeholder.com/48?text=Tutor';

              return (
                <AppointmentCard key={index} sx={{ position: 'relative' }}>
                  <Typography
                    variant="subtitle1"
                    sx={{ fontWeight: 'bold', mb: 1 }}
                  >
                    Appointment Date: {formattedDate}
                  </Typography>
                  <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                    Homework Completed: {percentage}
                  </Typography>
                  <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                    Duration: {displayDuration}
                  </Typography>
                  <Typography variant="body2" sx={{ color: '#333' }}>
                    Status: {status}
                  </Typography>

                  {/* Bottom-right Tutor Image & Name */}
                  <Box
                    sx={{
                      position: 'absolute',
                      bottom: 8, // Lower offset => visually lower in the card
                      right: 16,
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                    }}
                  >
                    <Avatar
                      src={tutorImage}
                      alt={appt.tutor || 'Tutor'}
                      sx={{ width: 48, height: 48, mb: 1 }}
                    />
                    <Typography variant="caption" sx={{ fontWeight: 'bold' }}>
                      {appt.tutor || 'Tutor'}
                    </Typography>
                  </Box>
                </AppointmentCard>
              );
            })
          ) : (
            <Typography variant="body2" color="textSecondary">
              No recent appointments available.
            </Typography>
          )}
        </Box>
      )}

      {/* Down Arrow */}
      <IconButton
        onClick={() => handleNextAppointment(sortedAppointments.length)}
        disabled={startIndex + 3 >= sortedAppointments.length}
        sx={{ mt: 2 }}
      >
        <KeyboardArrowDownIcon fontSize="large" />
      </IconButton>
    </Box>
  </SectionContainer>
</TabPanel>


          {/* ============== School Goals ============== */}
          <TabPanel value={activeTab} index={1}>
            <SectionContainer>
              <Divider sx={{ marginBottom: '16px' }} />

              {(studentData.goals || []).length > 0 ? (
                <TableContainer
                  sx={{
                    overflowX: {
                      xs: 'auto', // Scroll on small screens if needed
                      md: 'visible',
                    },
                  }}
                >
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell
                          align="center"
                          sx={{
                            fontWeight: 600,
                            whiteSpace: 'nowrap',
                          }}
                        >
                          School
                        </TableCell>
                        <TableCell
                          align="center"
                          sx={{
                            fontWeight: 600,
                            whiteSpace: 'nowrap',
                          }}
                        >
                          Percentile (25th, 50th, 75th)
                        </TableCell>
                      </TableRow>
                    </TableHead>

                    <TableBody>
                      {studentData.goals.map((goal, index) => {
                        const schoolName =
                          goal.university || goal.College || 'N/A';

                        // Ensure ACT_percentiles is joined by comma + space
                        const actPercentile =
                          goal.ACT_percentiles && goal.ACT_percentiles !== 'N/A'
                            ? Array.isArray(goal.ACT_percentiles)
                              ? goal.ACT_percentiles.join(', ')
                              : goal.ACT_percentiles
                            : 'No data';

                        // Ensure SAT_percentiles is joined by comma + space
                        const satPercentile =
                          goal.SAT_percentiles && goal.SAT_percentiles !== 'N/A'
                            ? Array.isArray(goal.SAT_percentiles)
                              ? goal.SAT_percentiles.join(', ')
                              : goal.SAT_percentiles
                            : 'No data';

                        return (
                          <TableRow key={index}>
                            <TableCell align="center">{schoolName}</TableCell>
                            <TableCell align="center">
                              <Box
                                sx={{
                                  display: 'flex',
                                  alignItems: 'center',
                                  justifyContent: 'center',
                                }}
                              >
                                <Box sx={{ mr: 2, textAlign: 'center' }}>
                                  <Typography variant="subtitle2">ACT</Typography>
                                  <Typography
                                    variant="body2"
                                    sx={{ whiteSpace: 'nowrap' }}
                                  >
                                    {actPercentile}
                                  </Typography>
                                </Box>

                                <Divider orientation="vertical" flexItem />

                                <Box sx={{ ml: 2, textAlign: 'center' }}>
                                  <Typography variant="subtitle2">SAT</Typography>
                                  <Typography
                                    variant="body2"
                                    sx={{ whiteSpace: 'nowrap' }}
                                  >
                                    {satPercentile}
                                  </Typography>
                                </Box>
                              </Box>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </TableContainer>
              ) : (
                <Typography variant="body2" color="textSecondary">
                  No school goals available.
                </Typography>
              )}
            </SectionContainer>
          </TabPanel>

          {/* ============== Student Profile ============== */}
          <TabPanel value={activeTab} index={2}>
            <SectionContainer>
            <SectionTitle variant="h6">Parent Profile</SectionTitle>
            <Divider sx={{ marginBottom: '16px' }} />
            {/* Bolded question prompt */}
            <Typography variant="body1" sx={{ fontWeight: 'bold', mb: 1 }}>
              Where should our invoices be emailed?
            </Typography>
            {/* Display current email when not editing or confirming */}
            {!editMode && (
              <Grid container alignItems="center" spacing={2}>
                <Grid item xs={8}>
                  <Typography>
                    {parentInvoiceEmail || 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={4}>
                  <Button variant="outlined" onClick={handleEditClick}>
                    Edit
                  </Button>
                </Grid>
              </Grid>
            )}
            {/* Edit mode view: show input field */}
            {editMode && !confirmationMode && (
              <Grid container alignItems="center" spacing={2} sx={{ marginTop: 2 }}>
                <Grid item xs={8}>
                  <TextField
                    fullWidth
                    variant="outlined"
                    value={invoiceEmailValue}
                    onChange={handleInputChange}
                    error={!!emailError}
                    helperText={emailError}
                  />
                </Grid>
                <Grid item xs={4}>
                  <Button variant="contained" onClick={handleSaveClick}>
                    Change
                  </Button>
                </Grid>
              </Grid>
            )}
            {/* Confirmation view */}
            {confirmationMode && (
              <Grid container direction="column" spacing={2} sx={{ marginTop: 2 }}>
                <Grid item>
                  <Typography variant="body1">
                    Please confirm that you want to update the invoice email to:{" "}
                    <strong>{invoiceEmailValue}</strong>
                  </Typography>
                </Grid>
                <Grid item container spacing={2}>
                  <Grid item>
                    <Button variant="contained" onClick={handleConfirmSave}>
                      Confirm
                    </Button>
                  </Grid>
                  <Grid item>
                    <Button variant="outlined" onClick={handleCancelConfirmation}>
                      Go Back
                    </Button>
                  </Grid>
                </Grid>
              </Grid>
            )}
          </SectionContainer>




  {/* Existing Student Profile Section */}
  <SectionContainer>
    <SectionTitle variant="h6">Student Profile</SectionTitle>
    <Divider sx={{ marginBottom: '16px' }} />
    <TableContainer>
      <Table>
        <TableBody>
          <TableRow>
            <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
            <TableCell>{studentData.personal?.name || 'N/A'}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell sx={{ fontWeight: 600 }}>Grade</TableCell>
            <TableCell>{studentData.personal?.grade || 'N/A'}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell sx={{ fontWeight: 600 }}>High School</TableCell>
            <TableCell>
              {studentData.personal?.high_school || 'N/A'}
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell sx={{ fontWeight: 600 }}>
              Accommodations
            </TableCell>
            <TableCell>
              {studentData.personal?.accommodations || 'N/A'}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </TableContainer>
  </SectionContainer>
</TabPanel>


 {/* ============== TEST DATA  ============== */}
          <TabPanel value={activeTab} index={3}>
            <Grid container spacing={4} direction={isMobile ? 'column' : 'row'}>
              {/* MAIN SECTION (Left/Top) */}
              <Grid item xs={12} md={10} order={{ xs: 1, md: 2 }}>
                <SectionContainer>
                  {/* NEW SUPER SCORE CHART REPLACING THE OLD CHART */}
                  <SuperScoreChart 
                    testData={testData} 
                    filter={testFocus === 'ACT' ? 'ACT' : 'SAT'} 
                  />

                  {/* SAT/PSAT Data Tables & Cards */}
                  {(testFocus === 'SAT' || testFocus === 'PSAT' || testFocus === 'TBD') && (
                    <>
                      <Typography variant="h6" sx={{ fontWeight: 600, mt: 2 }}>
                        SAT Test Results {testFocus === 'PSAT' && '(PSAT Student)'}
                      </Typography>
                      {/* Remove old chart – only tables/cards remain */}
                      {!isMobile ? (
                        <Box sx={{ mt: 4 }}>
                          <TableContainer>
                            <Table>
                              <TableHead>
                                <TableRow>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Date</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Test (Type)</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>EBRW</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Math</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Total</TableCell>
                                </TableRow>
                              </TableHead>
                              <TableBody>
                                {[...satTests]
                                  .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                                  .map((testDoc, index, arr) => {
                                    const { EBRW, Math, SAT_Total } = renderSATScores(testDoc);
                                    const prevDoc = arr[index + 1];
                                    const dateStr = testDoc.date || 'N/A';
                                    const label = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;
                                    return (
                                      <React.Fragment key={index}>
                                        <TableRow>
                                          <TableCell>{dateStr}</TableCell>
                                          <TableCell>{label}</TableCell>
                                          <TableCell>{EBRW}</TableCell>
                                          <TableCell>{Math}</TableCell>
                                          <TableCell>{SAT_Total}</TableCell>
                                        </TableRow>
                                        <TableRow>
                                          <TableCell colSpan={2} sx={{ fontWeight: 600 }}>
                                            {prevDoc ? 'Score Change' : 'Baseline'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.SAT_Scores?.[0]),
                                                parseScore(EBRW)
                                              ) : '—'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.SAT_Scores?.[1]),
                                                parseScore(Math)
                                              ) : '—'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.SAT_Scores?.[4]),
                                                parseScore(SAT_Total)
                                              ) : '—'}
                                          </TableCell>
                                        </TableRow>
                                        <TableRow>
                                          <TableCell colSpan={5} sx={{ px: 0 }}>
                                            <Divider sx={{ my: 2, borderColor: 'black', borderWidth: 2 }} />
                                          </TableCell>
                                        </TableRow>
                                      </React.Fragment>
                                    );
                                  })}
                              </TableBody>
                            </Table>
                          </TableContainer>
                        </Box>
                      ) : (
                        <Box sx={{ mt: 4 }}>
                          {[...satTests]
                            .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                            .map((testDoc, index, arr) => {
                              const { EBRW, Math, SAT_Total } = renderSATScores(testDoc);
                              const prevDoc = arr[index + 1];
                              const dateStr = testDoc.date || 'N/A';
                              const label = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;
                              return (
                                <Paper
                                  key={index}
                                  sx={{
                                    mb: 2,
                                    p: 2,
                                    borderRadius: '12px',
                                    boxShadow: '0 2px 12px rgba(0,0,0,0.1)',
                                  }}
                                >
                                  <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                                    {dateStr} — {label}
                                  </Typography>
                                  <Box
                                    sx={{
                                      display: 'flex',
                                      flexWrap: 'wrap',
                                      gap: 2,
                                      justifyContent: 'space-between',
                                    }}
                                  >
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        EBRW:
                                      </Typography>
                                      <Typography variant="body2">{EBRW}</Typography>
                                    </Box>
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        Math:
                                      </Typography>
                                      <Typography variant="body2">{Math}</Typography>
                                    </Box>
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        Total:
                                      </Typography>
                                      <Typography variant="body2">{SAT_Total}</Typography>
                                    </Box>
                                  </Box>
                                  {prevDoc ? (
                                    <Box sx={{ mt: 2 }}>
                                      <Typography variant="body2" sx={{ mb: 0.5, fontWeight: 'bold' }}>
                                        Score Change:
                                      </Typography>
                                      <Box
                                        sx={{
                                          display: 'flex',
                                          flexWrap: 'wrap',
                                          gap: 2,
                                          justifyContent: 'space-between',
                                        }}
                                      >
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            EBRW:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.SAT_Scores?.[0]),
                                              parseScore(EBRW)
                                            )}
                                          </Box>
                                        </Box>
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            Math:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.SAT_Scores?.[1]),
                                              parseScore(Math)
                                            )}
                                          </Box>
                                        </Box>
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            Total:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.SAT_Scores?.[4]),
                                              parseScore(SAT_Total)
                                            )}
                                          </Box>
                                        </Box>
                                      </Box>
                                    </Box>
                                  ) : (
                                    <Typography variant="body2" sx={{ mt: 2 }}>
                                      Baseline
                                    </Typography>
                                  )}
                                </Paper>
                              );
                            })}
                        </Box>
                      )}
                    </>
                  )}

                  {(testFocus === 'ACT' || testFocus === 'TBD') && (
                    <>
                      <Typography variant="h6" sx={{ fontWeight: 600, mt: 4 }}>
                        ACT Test Results
                      </Typography>
                      {/* Remove the old ACT chart; only display tables/cards */}
                      {!isMobile ? (
                        <Box sx={{ mt: 4 }}>
                          <TableContainer>
                            <Table>
                              <TableHead>
                                <TableRow>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Date</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Test (Type)</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>English</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Math</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Reading</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Science</TableCell>
                                  <TableCell sx={{ fontWeight: 'bold' }}>Total</TableCell>
                                </TableRow>
                              </TableHead>
                              <TableBody>
                                {[...actTests]
                                  .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                                  .map((testDoc, index, arr) => {
                                    const { English, MathVal, Reading, Science, ACT_Total } = renderACTScores(testDoc);
                                    const prevDoc = arr[index + 1];
                                    const dateStr = testDoc.date || 'N/A';
                                    const label = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;
                                    return (
                                      <React.Fragment key={index}>
                                        <TableRow>
                                          <TableCell>{dateStr}</TableCell>
                                          <TableCell>{label}</TableCell>
                                          <TableCell>{English}</TableCell>
                                          <TableCell>{MathVal}</TableCell>
                                          <TableCell>{Reading}</TableCell>
                                          <TableCell>{Science}</TableCell>
                                          <TableCell>{ACT_Total}</TableCell>
                                        </TableRow>
                                        <TableRow>
                                          <TableCell colSpan={2} sx={{ fontWeight: 600 }}>
                                            {prevDoc ? 'Score Change' : 'Baseline'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.ACT_Scores?.[0]),
                                                parseScore(English)
                                              ) : '—'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.ACT_Scores?.[1]),
                                                parseScore(MathVal)
                                              ) : '—'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.ACT_Scores?.[2]),
                                                parseScore(Reading)
                                              ) : '—'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.ACT_Scores?.[3]),
                                                parseScore(Science)
                                              ) : '—'}
                                          </TableCell>
                                          <TableCell>
                                            {prevDoc ? 
                                              renderChangeChip(
                                                parseScore(prevDoc.ACT_Scores?.[4]),
                                                parseScore(ACT_Total)
                                              ) : '—'}
                                          </TableCell>
                                        </TableRow>
                                        <TableRow>
                                          <TableCell colSpan={7} sx={{ px: 0 }}>
                                            <Divider sx={{ my: 2, borderColor: 'black', borderWidth: 2 }} />
                                          </TableCell>
                                        </TableRow>
                                      </React.Fragment>
                                    );
                                  })}
                              </TableBody>
                            </Table>
                          </TableContainer>
                        </Box>
                      ) : (
                        <Box sx={{ mt: 4 }}>
                          {[...actTests]
                            .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                            .map((testDoc, index, arr) => {
                              const { English, MathVal, Reading, Science, ACT_Total } = renderACTScores(testDoc);
                              const prevDoc = arr[index + 1];
                              const dateStr = testDoc.date || 'N/A';
                              const label = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;
                              return (
                                <Paper
                                  key={index}
                                  sx={{
                                    mb: 2,
                                    p: 2,
                                    borderRadius: '12px',
                                    boxShadow: '0 2px 12px rgba(0,0,0,0.1)',
                                  }}
                                >
                                  <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                                    {dateStr} — {label}
                                  </Typography>
                                  <Box
                                    sx={{
                                      display: 'flex',
                                      flexWrap: 'wrap',
                                      gap: 2,
                                      justifyContent: 'space-between',
                                    }}
                                  >
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        English:
                                      </Typography>
                                      <Typography variant="body2">{English}</Typography>
                                    </Box>
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        Math:
                                      </Typography>
                                      <Typography variant="body2">{MathVal}</Typography>
                                    </Box>
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        Reading:
                                      </Typography>
                                      <Typography variant="body2">{Reading}</Typography>
                                    </Box>
                                  </Box>
                                  {prevDoc ? (
                                    <Box sx={{ mt: 2 }}>
                                      <Typography variant="body2" sx={{ mb: 0.5, fontWeight: 'bold' }}>
                                        Score Change:
                                      </Typography>
                                      <Box
                                        sx={{
                                          display: 'flex',
                                          flexWrap: 'wrap',
                                          gap: 2,
                                          justifyContent: 'space-between',
                                        }}
                                      >
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            English:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.ACT_Scores?.[0]),
                                              parseScore(English)
                                            )}
                                          </Box>
                                        </Box>
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            Math:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.ACT_Scores?.[1]),
                                              parseScore(MathVal)
                                            )}
                                          </Box>
                                        </Box>
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            Reading:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.ACT_Scores?.[2]),
                                              parseScore(Reading)
                                            )}
                                          </Box>
                                        </Box>
                                      </Box>
                                    </Box>
                                  ) : (
                                    <Typography variant="body2" sx={{ mt: 2 }}>
                                      Baseline
                                    </Typography>
                                  )}
                                  <Divider sx={{ my: 2 }} />
                                  <Box
                                    sx={{
                                      display: 'flex',
                                      flexWrap: 'wrap',
                                      gap: 2,
                                      justifyContent: 'space-between',
                                    }}
                                  >
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        Science:
                                      </Typography>
                                      <Typography variant="body2">{Science}</Typography>
                                    </Box>
                                    <Box sx={{ flex: 1 }}>
                                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                        Total:
                                      </Typography>
                                      <Typography variant="body2">{ACT_Total}</Typography>
                                    </Box>
                                  </Box>
                                  {prevDoc ? (
                                    <Box sx={{ mt: 2 }}>
                                      <Typography variant="body2" sx={{ mb: 0.5, fontWeight: 'bold' }}>
                                        Score Change:
                                      </Typography>
                                      <Box
                                        sx={{
                                          display: 'flex',
                                          flexWrap: 'wrap',
                                          gap: 2,
                                          justifyContent: 'space-between',
                                        }}
                                      >
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            Science:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.ACT_Scores?.[3]),
                                              parseScore(Science)
                                            )}
                                          </Box>
                                        </Box>
                                        <Box sx={{ flex: 1 }}>
                                          <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                            Total:
                                          </Typography>
                                          <Box>
                                            {renderChangeChip(
                                              parseScore(prevDoc.ACT_Scores?.[4]),
                                              parseScore(ACT_Total)
                                            )}
                                          </Box>
                                        </Box>
                                      </Box>
                                    </Box>
                                  ) : (
                                    <Typography variant="body2" sx={{ mt: 2 }}>
                                      Baseline
                                    </Typography>
                                  )}
                                </Paper>
                              );
                            })}
                        </Box>
                      )}
                    </>
                  )}
                </SectionContainer>
              </Grid>

              {/* TEST DATES SIDE PANEL */}
              <Grid item xs={12} md={2} order={{ xs: 2, md: 1 }}>
                <SectionContainer>
                  <SectionTitle variant="h6">Testing Dates</SectionTitle>
                  <Divider sx={{ mb: 2 }} />
                  {(() => {
                    if (testFocus === 'TBD') {
                      return testDates.map((td, i) => (
                        <Box key={i} sx={{ mb: 2 }}>
                          <ListItemText
                            primary={td.test_date || 'N/A'}
                            secondary={td.test_type || 'N/A'}
                          />
                          <Divider sx={{ my: 1 }} />
                        </Box>
                      ));
                    } else if (testFocus === 'ACT') {
                      const relevant = testDates.filter((td) =>
                        (td.test_type || '').toUpperCase().includes('ACT')
                      );
                      return relevant.length ? (
                        relevant.map((td, i) => (
                          <Box key={i} sx={{ mb: 2 }}>
                            <ListItemText
                              primary={td.test_date || 'N/A'}
                              secondary={td.test_type || 'N/A'}
                            />
                            <Divider sx={{ my: 1 }} />
                          </Box>
                        ))
                      ) : (
                        <Typography variant="body2" color="textSecondary">
                          No upcoming ACT tests.
                        </Typography>
                      );
                    } else {
                      const relevant = testDates.filter((td) => {
                        const t = (td.test_type || '').toUpperCase();
                        return t.includes('SAT') || t.includes('PSAT');
                      });
                      return relevant.length ? (
                        relevant.map((td, i) => (
                          <Box key={i} sx={{ mb: 2 }}>
                            <ListItemText
                              primary={td.test_date || 'N/A'}
                              secondary={td.test_type || 'N/A'}
                            />
                            <Divider sx={{ my: 1 }} />
                          </Box>
                        ))
                      ) : (
                        <Typography variant="body2" color="textSecondary">
                          No upcoming SAT/PSAT tests.
                        </Typography>
                      );
                    }
                  })()}
                </SectionContainer>
              </Grid>
            </Grid>
          </TabPanel>


        </ContentWrapper>
      </Container>
    </RootContainer>
  );
};

export default ParentDashboard;