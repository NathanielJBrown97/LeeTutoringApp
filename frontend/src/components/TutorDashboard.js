import React, { useState, useContext, useEffect, forwardRef } from 'react';
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
  Pagination,
  Autocomplete,
  TextField,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Slide,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Checkbox,
  FormControlLabel,
  Slider
} from '@mui/material';
import { styled } from '@mui/system';
import TodaySchedule from './TodaySchedule';
import MySchedule from './MySchedule'; // NEW: Import MySchedule
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import collegeData from './collegeData'; // Import the college data
import CreateTestDataDialog from './createTestData';
import EditTestDataDialog from './editTestData';
import CreateHomeworkCompletionDialog from './createHomeworkCompletion';
import EditHomeworkCompletionDialog from './editHomeworkCompletion';

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

// New styled components for "My Students" section
const StyledStudentCard = styled(Card)(({ theme }) => ({
  backgroundColor: '#fff',
  border: `1px solid ${brandBlue}`,
  borderRadius: '8px',
  minHeight: '140px',
  display: 'flex',
  flexDirection: 'column',
  justifyContent: 'center',
  alignItems: 'center',
  padding: theme.spacing(2),
  cursor: 'pointer',
  transition: 'transform 0.3s, box-shadow 0.3s',
  '&:hover': {
    transform: 'scale(1.02)',
    boxShadow: '0 6px 16px rgba(0,0,0,0.15)'
  },
}));

const StyledExpandedStudentView = styled(Paper)(({ theme }) => ({
  padding: theme.spacing(3),
  borderRadius: '12px',
  backgroundColor: '#fff',
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
}));

// -------------------- New Dialog: EditBusinessDetailsDialog --------------------
function EditBusinessDetailsDialog({ open, onClose, onSubmit, initialData = {}, tutorFullName }) {
  // State for each field
  const [associatedTutors, setAssociatedTutors] = useState([]);
  const [scheduler, setScheduler] = useState('');
  const [status, setStatus] = useState('');
  const [teamLead, setTeamLead] = useState('');
  const [testFocus, setTestFocus] = useState('');
  const [registeredForTest, setRegisteredForTest] = useState(false);
  const [testDate, setTestDate] = useState('');
  const [notes, setNotes] = useState('');


  useEffect(() => {
    if (open && initialData) {
      setAssociatedTutors(initialData.associated_tutors || []);
      setScheduler(initialData.scheduler || '');
      setStatus(initialData.status || '');
      setTeamLead(initialData.team_lead || '');
      setTestFocus(initialData.test_focus || '');
      setRegisteredForTest(initialData.test_appointment ? initialData.test_appointment.registered_for_test : false);
      setTestDate(initialData.test_appointment ? initialData.test_appointment.test_date : '');
      setNotes(initialData.notes || '');
    }
  }, [open, initialData]);

  const handleSubmit = () => {
    const payload = {
      associated_tutors: associatedTutors,
      scheduler,
      status,
      team_lead: teamLead,
      test_focus: testFocus,
      test_appointment: {
        registered_for_test: registeredForTest,
        test_date: testDate,
      },
      notes, 
    };
    onSubmit(payload);
  };

  const associatedTutorsOptions = ['Edward', 'Kyra', 'Ben', 'Eli', 'Patrick', 'Kieran'];
  const schedulerOptions = ['Either Parent', 'Mother', 'Father', 'Student'];
  const statusOptions = ['Awaiting Results', 'Active', 'Inactive'];
  const teamLeadOptions = ['Edward', 'Eli', 'Ben', 'Kieran', 'Kyra', 'Patrick'];
  const testFocusOptions = ['ACT', 'SAT'];

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Business Details</DialogTitle>
      <DialogContent>
        <Box component="form" noValidate sx={{ mt: 2 }}>
          {/* Associated Tutors Multi-select */}
          <FormControl fullWidth margin="normal">
            <InputLabel id="associated-tutors-label">Associated Tutors</InputLabel>
            <Select
              labelId="associated-tutors-label"
              multiple
              value={associatedTutors}
              onChange={(e) => setAssociatedTutors(e.target.value)}
              label="Associated Tutors"
            >
              {associatedTutorsOptions.map((name) => (
                <MenuItem key={name} value={name}>
                  {name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          {/* Scheduler Dropdown */}
          <FormControl fullWidth margin="normal">
            <InputLabel id="scheduler-label">Scheduler</InputLabel>
            <Select
              labelId="scheduler-label"
              value={scheduler}
              label="Scheduler"
              onChange={(e) => setScheduler(e.target.value)}
            >
              {schedulerOptions.map((option) => (
                <MenuItem key={option} value={option}>{option}</MenuItem>
              ))}
            </Select>
          </FormControl>
          {/* Status Dropdown */}
          <FormControl fullWidth margin="normal">
            <InputLabel id="status-label">Status</InputLabel>
            <Select
              labelId="status-label"
              value={status}
              label="Status"
              onChange={(e) => setStatus(e.target.value)}
            >
              {statusOptions.map((option) => (
                <MenuItem key={option} value={option}>{option}</MenuItem>
              ))}
            </Select>
          </FormControl>
          {/* Team Lead Dropdown */}
          <FormControl fullWidth margin="normal">
            <InputLabel id="team-lead-label">Team Lead</InputLabel>
            <Select
              labelId="team-lead-label"
              value={teamLead}
              label="Team Lead"
              onChange={(e) => setTeamLead(e.target.value)}
            >
              {teamLeadOptions.map((option) => (
                <MenuItem key={option} value={option}>{option}</MenuItem>
              ))}
            </Select>
          </FormControl>
          {/* Test Focus Dropdown */}
          <FormControl fullWidth margin="normal">
            <InputLabel id="test-focus-label">Test Focus</InputLabel>
            <Select
              labelId="test-focus-label"
              value={testFocus}
              label="Test Focus"
              onChange={(e) => setTestFocus(e.target.value)}
            >
              {testFocusOptions.map((option) => (
                <MenuItem key={option} value={option}>{option}</MenuItem>
              ))}
            </Select>
          </FormControl>
          {/* Test Appointment */}
          <Box sx={{ display: 'flex', alignItems: 'center', mt: 2 }}>
            <FormControl margin="normal" sx={{ mr: 2 }}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={registeredForTest}
                    onChange={(e) => setRegisteredForTest(e.target.checked)}
                  />
                }
                label="Registered for Test"
              />
            </FormControl>
            <TextField
              label="Test Date"
              type="date"
              value={testDate}
              onChange={(e) => setTestDate(e.target.value)}
              InputLabelProps={{ shrink: true }}
              margin="normal"
            />
            <TextField
              label="Notes"
              type="text"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              fullWidth
              margin="normal"
            />

          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit} variant="contained">Save Changes</Button>
      </DialogActions>
    </Dialog>
  );
}

// -------------------- EditNotesDialog Component --------------------
const EditNotesDialog = ({ open, onClose, onSubmit, initialNotes = '' }) => {
  const [notes, setNotes] = useState(initialNotes);

  useEffect(() => {
    setNotes(initialNotes);
  }, [initialNotes]);

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Update Notes:</DialogTitle>
      <DialogContent>
        <TextField
          autoFocus
          margin="dense"
          label="Notes"
          type="text"
          fullWidth
          variant="outlined"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={() => onSubmit(notes)} variant="contained">
          Submit
        </Button>
      </DialogActions>
    </Dialog>
  );
};

// -------------------- Simple TabPanel helper --------------------
function TabPanel(props) {
  const { children, value, index, ...other } = props;
  return (
    <Box role="tabpanel" hidden={value !== index} {...other} sx={{ marginTop: '16px' }}>
      {value === index && children}
    </Box>
  );
}

// -------------------- Updated InfoCard Component --------------------
function InfoCard({ item, onEdit = (item) => { console.log("Edit not implemented", item); }, onDelete = (item) => { console.log("Delete not implemented", item); } }) {
  return (
    <Card variant="outlined" sx={{ marginBottom: 2 }}>
      <CardContent>
        {Object.entries(item).map(([key, value]) => (
          <Typography key={key} variant="body2">
            <strong>{key}:</strong> {String(value)}
          </Typography>
        ))}
      </CardContent>
      <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 0.5, p: 1 }}>
        {onEdit != null && (
          <Button variant="text" size="small" onClick={() => onEdit(item)}>
            <EditIcon fontSize="small" />
          </Button>
        )}
        <Button variant="text" size="small" onClick={() => onDelete(item)}>
          <DeleteIcon fontSize="small" />
        </Button>
      </Box>
    </Card>
  );
}

// -------------------- NewGoalDialog Component --------------------
const Transition = forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

function NewGoalDialog({ open, onClose, onSubmit }) {
  const [selectedCollege, setSelectedCollege] = useState(null);

  return (
    <Dialog open={open} TransitionComponent={Transition} keepMounted onClose={onClose}>
      <DialogTitle>Create New Goal</DialogTitle>
      <DialogContent>
        <Autocomplete
          options={collegeData}
          getOptionLabel={(option) => option.school}
          renderInput={(params) => <TextField {...params} label="Select College" variant="outlined" />}
          onChange={(event, value) => setSelectedCollege(value)}
          fullWidth
          clearOnEscape
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button
          onClick={() => {
            if (selectedCollege) {
              onSubmit(selectedCollege);
            }
          }}
          variant="contained"
        >
          Submit
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function EditPersonalDetailsDialog({ open, onClose, onSubmit, initialData = {} }) {
  const [name, setName] = useState(initialData.name || '');
  const [accommodations, setAccommodations] = useState(initialData.accommodations || '');
  const [grade, setGrade] = useState(initialData.grade || '');
  const [highSchool, setHighSchool] = useState(initialData.high_school || '');
  const [parentEmail, setParentEmail] = useState(initialData.parent_email || '');
  const [studentEmail, setStudentEmail] = useState(initialData.student_email || '');
  const [interests, setInterests] = useState(initialData.interests || '');
  const [parentNumber, setParentNumber] = useState(initialData.parent_number || '');
  const [studentNumber, setStudentNumber] = useState(initialData.student_number || '');

  useEffect(() => {
    setName(initialData.name || '');
    setAccommodations(initialData.accommodations || '');
    setGrade(initialData.grade || '');
    setHighSchool(initialData.high_school || '');
    setParentEmail(initialData.parent_email || '');
    setStudentEmail(initialData.student_email || '');
    setInterests(initialData.interests || '');
    setParentNumber(initialData.parent_number || '');
    setStudentNumber(initialData.student_number || '');
  }, [initialData]);

  const handleSubmit = () => {
    const payload = {
      name,
      accommodations,
      grade,
      high_school: highSchool,
      parent_email: parentEmail,
      student_email: studentEmail,
      interests,
      parent_number: parentNumber,
      student_number: studentNumber,
    };
    onSubmit(payload);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Personal Details</DialogTitle>
      <DialogContent>
        <TextField
          margin="normal"
          label="Name"
          fullWidth
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <TextField
          margin="normal"
          label="Accommodations"
          fullWidth
          value={accommodations}
          onChange={(e) => setAccommodations(e.target.value)}
        />
        <TextField
          margin="normal"
          label="Grade"
          fullWidth
          value={grade}
          onChange={(e) => setGrade(e.target.value)}
        />
        <TextField
          margin="normal"
          label="High School"
          fullWidth
          value={highSchool}
          onChange={(e) => setHighSchool(e.target.value)}
        />
        <TextField
          margin="normal"
          label="Parent Email"
          fullWidth
          value={parentEmail}
          onChange={(e) => setParentEmail(e.target.value)}
        />
        <TextField
          margin="normal"
          label="Student Email"
          fullWidth
          value={studentEmail}
          onChange={(e) => setStudentEmail(e.target.value)}
        />
        <TextField
          margin="normal"
          label="Interests"
          fullWidth
          value={interests}
          onChange={(e) => setInterests(e.target.value)}
        />
        <TextField
          margin="normal"
          label="Parent Number"
          fullWidth
          value={parentNumber}
          onChange={(e) => setParentNumber(e.target.value)}
        />
        <TextField
          margin="normal"
          label="Student Number"
          fullWidth
          value={studentNumber}
          onChange={(e) => setStudentNumber(e.target.value)}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit} variant="contained">
          Save Changes
        </Button>
      </DialogActions>
    </Dialog>
  );
}

// -------------------- StudentsTab Component --------------------
function StudentsTab({ tutorId, tutorEmail, backendUrl, filterTodayAppointments = false, enableSearch = false, tutorName }) {
  const [studentIds, setStudentIds] = useState([]);
  const [studentDetails, setStudentDetails] = useState([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedStudent, setSelectedStudent] = useState(null);
  const [innerTab, setInnerTab] = useState(0);
  const [showNewGoalDialog, setShowNewGoalDialog] = useState(false);
  const [openTestDatesDialog, setOpenTestDatesDialog] = useState(false);
  const [currentTestDate, setCurrentTestDate] = useState(null);
  const [showNewTestDataDialog, setShowNewTestDataDialog] = useState(false);
  const [showEditTestDataDialog, setShowEditTestDataDialog] = useState(false);
  const [editingTestData, setEditingTestData] = useState(null);
  const [showNewHomeworkDialog, setShowNewHomeworkDialog] = useState(false);
  const [showEditHomeworkDialog, setShowEditHomeworkDialog] = useState(false);
  const [editingHomework, setEditingHomework] = useState(null);
  const [showEditBusinessDialog, setShowEditBusinessDialog] = useState(false);
  const [showEditPersonalDialog, setShowEditPersonalDialog] = useState(false);
  const [editingPersonal, setEditingPersonal] = useState(null);

  const studentsPerPage = 10;
  const theme = useTheme();
  const isSmallScreen = useMediaQuery(theme.breakpoints.down('sm'));

  const hadAppointmentToday = (student) => {
    if (!student.appointments || student.appointments.length === 0) return false;
    const today = new Date();
    return student.appointments.some((appointment) => {
      const appDate = new Date(appointment.date);
      return (
        appDate.getFullYear() === today.getFullYear() &&
        appDate.getMonth() === today.getMonth() &&
        appDate.getDate() === today.getDate()
      );
    });
  };
  useEffect(() => {
    if (selectedStudent && selectedStudent.id) {
      const token = localStorage.getItem('authToken');
      fetch(
        `${backendUrl}/api/tutor/get-personal-details?firebase_id=${encodeURIComponent(selectedStudent.id)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        }
      )
        .then((res) => {
          if (!res.ok) {
            throw new Error(`Failed to fetch personal details. Status: ${res.status}`);
          }
          return res.json();
        })
        .then((data) => {
          setSelectedStudent((prev) => ({
            ...prev,
            personal: data.personal,
          }));
        })
        .catch((error) => {
          console.error("Error fetching personal details:", error);
        });
    }
  }, [selectedStudent?.id, backendUrl]);
  
  useEffect(() => {
    if (selectedStudent && selectedStudent.id) {
      const token = localStorage.getItem('authToken');
      fetch(`${backendUrl}/api/tutor/get-business-details?firebase_id=${encodeURIComponent(selectedStudent.id)}`, {
        method: 'GET',
        headers: {
           'Content-Type': 'application/json',
           'Authorization': `Bearer ${token}`,
        },
      })
        .then(res => {
          if (!res.ok) {
            throw new Error(`Failed to fetch business details. Status: ${res.status}`);
          }
          return res.json();
        })
        .then(data => {
          setSelectedStudent(prev => ({
            ...prev,
            business: data.business,
          }));
        })
        .catch(error => {
          console.error("Error fetching business details:", error);
        });
    }
  }, [selectedStudent?.id, backendUrl]);
  
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
        const ids = await res.json();
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

  const filteredDetails = filterTodayAppointments ? studentDetails.filter(hadAppointmentToday) : studentDetails;
  const indexOfLast = currentPage * studentsPerPage;
  const indexOfFirst = indexOfLast - studentsPerPage;
  const currentStudents = filteredDetails.slice(indexOfFirst, indexOfLast);
  const totalPages = Math.ceil(filteredDetails.length / studentsPerPage);

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

  const handleCreateNew = (tabIndex) => {
    if (tabIndex === 3) {
      setShowNewGoalDialog(true);
    } else if (tabIndex === 1) {
      setShowNewTestDataDialog(true);
    } else if (tabIndex === 0) {
      setShowNewHomeworkDialog(true);
    } else {
      console.log(`Creating new entry for tab index ${tabIndex}`);
    }
  };

  const handleNewGoalSubmit = async (college) => {
    const payload = {
      firebase_id: selectedStudent.id,
      college: college.school,
      act_percentiles: {
        p25: college.act25,
        p50: college.act50,
        p75: college.act75
      },
      sat_percentiles: {
        p25: college.sat25,
        p50: college.sat50,
        p75: college.sat75
      }
    };
    console.log("Submitting new goal payload:", payload);
    try {
      const token = localStorage.getItem('authToken');
      const res = await fetch(`${backendUrl}/api/tutor/create-goal`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        console.error('Failed to create new goal. Status:', res.status);
      } else {
        console.log('New goal created successfully.');
        const res2 = await fetch(
          `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
          {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
          }
        );
        if (res2.ok) {
          const updatedStudent = await res2.json();
          setSelectedStudent(updatedStudent);
        } else {
          console.error('Failed to refresh student details. Status:', res2.status);
        }
      }
    } catch (error) {
      console.error('Error creating new goal:', error);
    } finally {
      setShowNewGoalDialog(false);
    }
  };

  const handleEditHomework = (item) => {
    setEditingHomework(item);
    setShowEditHomeworkDialog(true);
  };

  const handleDeleteHomeworkCompletion = async (item) => {
    const confirmDelete = window.confirm("Are you sure you want to delete this homework completion record?");
    if (!confirmDelete) return;
    try {
      const token = localStorage.getItem('authToken');
      const res = await fetch(`${backendUrl}/api/tutor/delete-homework-completion`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          firebase_id: selectedStudent.id,
          event_id: item.id,
        }),
      });
      if (!res.ok) {
        console.error('Failed to delete homework completion. Status:', res.status);
        alert("Failed to delete homework completion record.");
        return;
      }
      alert("Homework completion record deleted successfully.");
      const res2 = await fetch(
        `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        }
      );
      if (res2.ok) {
        const updatedStudent = await res2.json();
        setSelectedStudent(updatedStudent);
      }
    } catch (error) {
      console.error('Error deleting homework completion record:', error);
      alert("Error deleting homework completion record.");
    }
  };

  const handleEditTestData = (item) => {
    setEditingTestData(item);
    setShowEditTestDataDialog(true);
  };

  const handleEditTestDataSubmit = (payload) => {
    const token = localStorage.getItem('authToken');
    fetch(`${backendUrl}/api/tutor/edit-test-data`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify({
        firebase_id: selectedStudent.id,
        ...payload,
      }),
    })
      .then((res) => {
        if (!res.ok) {
          alert("Failed to edit test data.");
          throw new Error("Failed to edit test data.");
        } else {
          alert("Test data updated successfully.");
          return fetch(
            `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
            {
              method: 'GET',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
              },
            }
          );
        }
      })
      .then((res2) => {
        if (!res2.ok) {
          throw new Error("Failed to refresh student details.");
        }
        return res2.json();
      })
      .then((updatedStudent) => {
        setSelectedStudent(updatedStudent);
      })
      .catch((error) => {
        console.error('Error editing test data:', error);
        alert("Error editing test data.");
      })
      .finally(() => {
        setShowEditTestDataDialog(false);
        setEditingTestData(null);
      });
  };

  const handleDeleteTestData = async (item) => {
    const confirmDelete = window.confirm("Are you sure you want to delete this test data?");
    if (!confirmDelete) return;
  
    try {
      const token = localStorage.getItem('authToken');
      const res = await fetch(`${backendUrl}/api/tutor/delete-test-data`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          firebase_id: selectedStudent.id,
          test_data_id: item.id,
        }),
      });
      if (!res.ok) {
        console.error('Failed to delete test data. Status:', res.status);
        alert("Failed to delete test data.");
        return;
      }
      alert("Test data deleted successfully.");
      const res2 = await fetch(
        `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        }
      );
      if (res2.ok) {
        const updatedStudent = await res2.json();
        setSelectedStudent(updatedStudent);
      }
    } catch (error) {
      console.error('Error deleting test data:', error);
      alert("Error deleting test data.");
    }
  };

  const handleDeleteGoal = async (goal) => {
    const confirmDelete = window.confirm("Are you sure you want to delete this goal?");
    if (!confirmDelete) return;
    const payload = {
      firebase_id: selectedStudent.id,
      college: goal.College,
    };
    try {
      const token = localStorage.getItem('authToken');
      const res = await fetch(`${backendUrl}/api/tutor/delete-goal`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        console.error('Failed to delete goal. Status:', res.status);
        alert("Failed to delete goal.");
        return;
      }
      alert("Goal deleted successfully.");
      const res2 = await fetch(
        `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        }
      );
      if (res2.ok) {
        const updatedStudent = await res2.json();
        setSelectedStudent(updatedStudent);
      } else {
        console.error('Failed to refresh student details. Status:', res2.status);
      }
    } catch (error) {
      console.error('Error deleting goal:', error);
      alert("Error deleting goal.");
    }
  };

  const handleDeleteEvent = async (item) => {
    const confirmDelete = window.confirm("Are you sure you want to delete this event?");
    if (!confirmDelete) return;
    try {
      const token = localStorage.getItem('authToken');
      const res = await fetch(`${backendUrl}/api/tutor/delete-event`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          firebase_id: selectedStudent.id,
          event_id: item.id,
        }),
      });
      if (!res.ok) {
        console.error('Failed to delete event. Status:', res.status);
        alert("Failed to delete event.");
        return;
      }
      alert("Event deleted successfully.");
      const res2 = await fetch(
        `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        }
      );
      if (res2.ok) {
        const updatedStudent = await res2.json();
        setSelectedStudent(updatedStudent);
      }
    } catch (error) {
      console.error('Error deleting event:', error);
      alert("Error deleting event.");
    }
  };

  const getCreateButtonLabel = (tabIndex) => {
    switch(tabIndex) {
      case 0: return 'Create New Homework';
      case 1: return 'Create New Test Data';
      case 2: return 'Update Test Dates';
      case 3: return 'Create New Goal';
      default: return 'Create New';
    }
  };


  const handleEditPersonalDetails = () => {
    if (!selectedStudent) return;
    // Initialize the dialog fields with the currently stored personal details.
    setEditingPersonal(selectedStudent.personal);
    setShowEditPersonalDialog(true);
  };

  // Updated: Open the Edit Business Details dialog
  const handleEditBusinessDetails = () => {
    setShowEditBusinessDialog(true);
  };

  if (selectedStudent) {
    const sortedHomework = selectedStudent.homeworkCompletion && selectedStudent.homeworkCompletion.length > 0
      ? selectedStudent.homeworkCompletion.slice().sort((a, b) => {
          if(a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
          if(a.date && b.date) return new Date(b.date) - new Date(a.date);
          return 0;
        })
      : [];
    const sortedTestData = selectedStudent.testData && selectedStudent.testData.length > 0
      ? selectedStudent.testData.slice().sort((a, b) => {
          if(a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
          if(a.date && b.date) return new Date(b.date) - new Date(a.date);
          return 0;
        })
      : [];
    const sortedTestDates = selectedStudent.testDates && selectedStudent.testDates.length > 0
      ? selectedStudent.testDates.slice().sort((a, b) => {
          if(a.test_date && b.test_date) return new Date(b.test_date) - new Date(a.test_date);
          return 0;
        })
      : [];
    const sortedGoals = selectedStudent.goals && selectedStudent.goals.length > 0
      ? selectedStudent.goals.slice().sort((a, b) => {
          if(a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
          if(a.date && b.date) return new Date(b.date) - new Date(a.date);
          return 0;
        })
      : [];

    const ExpandedViewContent = (
      <Box sx={{ marginBottom: 2 }}>
        <Typography variant="h5" sx={{ marginBottom: 2 }}>
          {selectedStudent.personal?.name || 'Student Overview'}
        </Typography>
        <Box sx={{ marginBottom: 2 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <Typography variant="subtitle1">Personal Details:</Typography>
                    <IconButton size="small" onClick={handleEditPersonalDetails}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                  </Box>
                  <Typography>
                    <strong>Name:</strong> {selectedStudent.personal?.name || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>Accommodations:</strong> {selectedStudent.personal?.accommodations || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>Grade:</strong> {selectedStudent.personal?.grade || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>High School:</strong> {selectedStudent.personal?.high_school || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>Parent Email:</strong> {selectedStudent.personal?.parent_email || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>Student Email:</strong> {selectedStudent.personal?.student_email || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>Interests:</strong> {selectedStudent.personal?.interests || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>Parent Number:</strong> {selectedStudent.personal?.parent_number || 'N/A'}
                  </Typography>
                  <Typography>
                    <strong>Student Number:</strong> {selectedStudent.personal?.student_number || 'N/A'}
                  </Typography>
                </Box>
      <Box sx={{ marginBottom: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <Typography variant="subtitle1">Business Details:</Typography>
        <IconButton size="small" onClick={handleEditBusinessDetails}>
          <EditIcon fontSize="small" />
        </IconButton>
      </Box>
      <Typography>
        <strong>Associated Tutors:</strong> {selectedStudent.business?.associated_tutors ? selectedStudent.business.associated_tutors.join(', ') : 'N/A'}
      </Typography>
      <Typography>
        <strong>Scheduler:</strong> {selectedStudent.business?.scheduler || 'N/A'}
      </Typography>
      <Typography>
        <strong>Status:</strong> {selectedStudent.business?.status || 'N/A'}
      </Typography>
      <Typography>
        <strong>Team Lead:</strong> {selectedStudent.business?.team_lead || 'N/A'}
      </Typography>
      <Typography>
        <strong>Test Focus:</strong> {selectedStudent.business?.test_focus || 'N/A'}
      </Typography>
      <Typography>
        <strong>Test Appointment - Registered:</strong> {selectedStudent.business?.test_appointment?.registered_for_test ? 'Yes' : 'No'}
      </Typography>
      <Typography>
        <strong>Test Appointment - Date:</strong> {selectedStudent.business?.test_appointment?.test_date || 'N/A'}
      </Typography>
      <Typography>
        <strong>Notes:</strong> {selectedStudent.business?.notes || 'N/A'}
      </Typography>
    </Box>


        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 2 }}>
          <Tabs value={innerTab} onChange={(e, newValue) => setInnerTab(newValue)}>
            <Tab label="Homework Completion" />
            <Tab label="Test Data" />
            <Tab label="Test Dates" />
            <Tab label="Goals" />
            <Tab label="Events" />
          </Tabs>
          <Button
            variant="contained"
            size={isSmallScreen ? "small" : "medium"}
            onClick={() => handleCreateNew(innerTab)}
            sx={{ marginLeft: 2 }}
          >
            {getCreateButtonLabel(innerTab)}
          </Button>
        </Box>
        <TabPanel value={innerTab} index={0}>
          {sortedHomework.length > 0 ? (
            sortedHomework.map((item) => (
              <InfoCard key={item.id} item={item} onEdit={handleEditHomework} onDelete={handleDeleteHomeworkCompletion} />
            ))
          ) : (
            <Typography>No Homework Completion data available.</Typography>
          )}
        </TabPanel>
        <TabPanel value={innerTab} index={1}>
          {sortedTestData.length > 0 ? (
            sortedTestData.map((item) => (
              <Card key={item.id} variant="outlined" sx={{ marginBottom: 2, padding: 2 }}>
                <Typography variant="subtitle1" sx={{ mb: 1 }}>
                  Test Data: {item.id}
                </Typography>
                <Typography>
                  <strong>Date:</strong> {item.date}
                </Typography>
                <Typography>
                  <strong>Baseline:</strong> {String(item.baseline)}
                </Typography>
                <Typography>
                  <strong>Test:</strong> {item.test}
                </Typography>
                <Typography>
                  <strong>Type:</strong> {item.type}
                </Typography>
                {item.ACT_Scores && (
                  <Box sx={{ mt: 1 }}>
                    <Typography variant="subtitle2">ACT Scores:</Typography>
                    <Typography>
                      <strong>ACT_Total:</strong> {item.ACT_Scores.ACT_Total}
                    </Typography>
                    <Typography>
                      <strong>English:</strong> {item.ACT_Scores.English}
                    </Typography>
                    <Typography>
                      <strong>Math:</strong> {item.ACT_Scores.Math}
                    </Typography>
                    <Typography>
                      <strong>Reading:</strong> {item.ACT_Scores.Reading}
                    </Typography>
                    <Typography>
                      <strong>Science:</strong> {item.ACT_Scores.Science}
                    </Typography>
                  </Box>
                )}
                {item.SAT_Scores && (
                  <Box sx={{ mt: 1 }}>
                    <Typography variant="subtitle2">SAT Scores:</Typography>
                    <Typography>
                      <strong>SAT_Total:</strong> {item.SAT_Scores.SAT_Total}
                    </Typography>
                    <Typography>
                      <strong>EBRW:</strong> {item.SAT_Scores.EBRW}
                    </Typography>
                    <Typography>
                      <strong>Math:</strong> {item.SAT_Scores.Math}
                    </Typography>
                    <Typography>
                      <strong>Reading:</strong> {item.SAT_Scores.Reading}
                    </Typography>
                    <Typography>
                      <strong>Writing:</strong> {item.SAT_Scores.Writing}
                    </Typography>
                  </Box>
                )}
                <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 0.5, p: 1 }}>
                  <Button variant="text" size="small" onClick={() => handleEditTestData(item)}>
                    <EditIcon fontSize="small" />
                  </Button>
                  <Button variant="text" size="small" onClick={() => handleDeleteTestData(item)}>
                    <DeleteIcon fontSize="small" />
                  </Button>
                </Box>
              </Card>
            ))
          ) : (
            <Typography>No Test Data available.</Typography>
          )}
        </TabPanel>
        <TabPanel value={innerTab} index={2}>
          {sortedTestDates.length > 0 ? (
            sortedTestDates.map((item) => (
              <Card key={item.id} variant="outlined" sx={{ marginBottom: 2 }}>
                <CardContent>
                  {Object.entries(item).map(([key, value]) => (
                    <Typography key={key} variant="body2">
                      <strong>{key}:</strong> {String(value)}
                    </Typography>
                  ))}
                </CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'flex-end', p: 1 }}>
                  <Button variant="text" size="small" onClick={() => {
                    setCurrentTestDate(item);
                    setOpenTestDatesDialog(true);
                  }}>
                    <EditIcon fontSize="small" />
                  </Button>
                </Box>
              </Card>
            ))
          ) : (
            <Typography>No Test Dates available.</Typography>
          )}
        </TabPanel>
        <TabPanel value={innerTab} index={3}>
          {sortedGoals.length > 0 ? (
            sortedGoals.map((item) => (
              <InfoCard key={item.id} item={item} onDelete={handleDeleteGoal} onEdit={enableSearch ? null : undefined} />
            ))
          ) : (
            <Typography>No Goals available.</Typography>
          )}
        </TabPanel>
        <TabPanel value={innerTab} index={4}>
          {selectedStudent.events && selectedStudent.events.length > 0 ? (
            selectedStudent.events.map((item) => (
              <InfoCard key={item.id} item={item} onDelete={handleDeleteEvent} />
            ))
          ) : (
            <Typography>No Events available.</Typography>
          )}
        </TabPanel>
        {showNewGoalDialog && (
          <NewGoalDialog
            open={showNewGoalDialog}
            onClose={() => setShowNewGoalDialog(false)}
            onSubmit={handleNewGoalSubmit}
          />
        )}
        {showNewTestDataDialog && (
          <CreateTestDataDialog
            open={showNewTestDataDialog}
            onClose={() => setShowNewTestDataDialog(false)}
            onSubmit={(payload) => {
              const newPayload = {
                ...payload,
                firebase_id: selectedStudent.id,
              };
              console.log("New Test Data payload:", newPayload);
              const token = localStorage.getItem('authToken');
              fetch(`${backendUrl}/api/tutor/create-test-data`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify(newPayload),
              })
                .then((res) => {
                  if (!res.ok) {
                    console.error('Failed to create test data. Status:', res.status);
                    alert("Failed to create test data.");
                  } else {
                    alert("Test data created successfully.");
                    fetch(
                      `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
                      {
                        method: 'GET',
                        headers: {
                          'Content-Type': 'application/json',
                          'Authorization': `Bearer ${token}`,
                        },
                      }
                    )
                      .then((res2) => {
                        if (res2.ok) {
                          return res2.json();
                        }
                        throw new Error(`Failed to refresh student details. Status: ${res2.status}`);
                      })
                      .then((updatedStudent) => {
                        setSelectedStudent(updatedStudent);
                      })
                      .catch((error) => {
                        console.error(error);
                      });
                  }
                })
                .catch((error) => {
                  console.error('Error creating test data:', error);
                  alert("Error creating test data.");
                })
                .finally(() => {
                  setShowNewTestDataDialog(false);
                });
            }}
          />
        )}
        {showNewHomeworkDialog && (
          <CreateHomeworkCompletionDialog
            open={showNewHomeworkDialog}
            onClose={() => setShowNewHomeworkDialog(false)}
            onSubmit={(payload) => {
              const token = localStorage.getItem('authToken');
              fetch(`${backendUrl}/api/tutor/create-homework-completion`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify({
                  firebase_id: selectedStudent.id,
                  ...payload
                }),
              })
                .then((res) => {
                  if (!res.ok) {
                    alert("Failed to create homework completion record.");
                  } else {
                    alert("Homework completion record created successfully.");
                    return fetch(
                      `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
                      {
                        method: 'GET',
                        headers: {
                          'Content-Type': 'application/json',
                          'Authorization': `Bearer ${token}`,
                        },
                      }
                    );
                  }
                })
                .then((res2) => res2 && res2.json())
                .then((updatedStudent) => {
                  updatedStudent && setSelectedStudent(updatedStudent);
                })
                .catch((error) => console.error(error))
                .finally(() => setShowNewHomeworkDialog(false));
            }}
            tutorFullName={tutorName || ''}
          />
        )}
        {showEditHomeworkDialog && editingHomework && (
          <EditHomeworkCompletionDialog
            open={showEditHomeworkDialog}
            onClose={() => {
              setShowEditHomeworkDialog(false);
              setEditingHomework(null);
            }}
            onSubmit={(payload) => {
              const token = localStorage.getItem('authToken');
              fetch(`${backendUrl}/api/tutor/edit-homework-completion`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify({
                  firebase_id: selectedStudent.id,
                  ...payload
                }),
              })
                .then((res) => {
                  if (!res.ok) {
                    alert("Failed to update homework completion record.");
                  } else {
                    alert("Homework completion record updated successfully.");
                    return fetch(
                      `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
                      {
                        method: 'GET',
                        headers: {
                          'Content-Type': 'application/json',
                          'Authorization': `Bearer ${token}`,
                        },
                      }
                    );
                  }
                })
                .then((res2) => res2 && res2.json())
                .then((updatedStudent) => {
                  updatedStudent && setSelectedStudent(updatedStudent);
                })
                .catch((error) => console.error(error))
                .finally(() => {
                  setShowEditHomeworkDialog(false);
                  setEditingHomework(null);
                });
            }}
            tutorFullName={tutorName || ''}
            initialData={editingHomework}
          />
        )}
        {showEditTestDataDialog && editingTestData && (
          <EditTestDataDialog
            open={showEditTestDataDialog}
            onClose={() => {
              setShowEditTestDataDialog(false);
              setEditingTestData(null);
            }}
            onSubmit={handleEditTestDataSubmit}
            initialData={editingTestData}
          />
        )}
        {openTestDatesDialog && (
          <EditNotesDialog
            open={openTestDatesDialog}
            onClose={() => {
              setOpenTestDatesDialog(false);
              setCurrentTestDate(null);
            }}
            onSubmit={(notes) => {
              const payload = {
                firebase_id: selectedStudent.id,
                document_name: currentTestDate.id,
                notes: notes,
              };
              fetch(`${backendUrl}/api/tutor/edit-test-dates-notes`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
                },
                body: JSON.stringify(payload),
              })
                .then((res) => {
                  if (!res.ok) {
                    console.error('Failed to update test date notes. Status:', res.status);
                    alert("Failed to update notes.");
                  } else {
                    alert("Test date notes updated successfully.");
                    const updatedTestDates = selectedStudent.testDates.map((td) =>
                      td.id === currentTestDate.id ? { ...td, notes: payload.notes } : td
                    );
                    setSelectedStudent({ ...selectedStudent, testDates: updatedTestDates });
                  }
                })
                .catch((error) => {
                  console.error('Error updating test date notes:', error);
                  alert("Error updating notes.");
                })
                .finally(() => {
                  setOpenTestDatesDialog(false);
                  setCurrentTestDate(null);
                });
            }}
            initialNotes={currentTestDate ? currentTestDate.notes || '' : ''}
          />
        )}

        {showEditPersonalDialog && (
        <EditPersonalDetailsDialog
          open={showEditPersonalDialog}
          onClose={() => {
            setShowEditPersonalDialog(false);
            setEditingPersonal(null);
          }}
          onSubmit={(payload) => {
            const token = localStorage.getItem('authToken');
            // Call the endpoint to update the personal details.
            fetch(`${backendUrl}/api/tutor/edit-personal-details`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
              },
              body: JSON.stringify({ firebase_id: selectedStudent.id, ...payload })
            })
              .then((res) => {
                if (!res.ok) {
                  throw new Error("Failed to update personal details.");
                }
                // After updating, fetch the student record again.
                return fetch(`${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`, {
                  method: 'GET',
                  headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                  },
                });
              })
              .then((res2) => {
                if (!res2.ok) {
                  throw new Error("Failed to refresh student details.");
                }
                return res2.json();
              })
              .then((updatedStudent) => {
                setSelectedStudent(updatedStudent);
              })
              .catch((error) => {
                console.error("Error updating personal details:", error);
                alert("Error updating personal details");
              })
              .finally(() => {
                setShowEditPersonalDialog(false);
              });
          }}
          initialData={editingPersonal}
        />
      )}

        {/* NEW: Render the Edit Business Details dialog */}
        {showEditBusinessDialog && (
          <EditBusinessDetailsDialog
            open={showEditBusinessDialog}
            onClose={() => setShowEditBusinessDialog(false)}
            initialData={selectedStudent.business}
            onSubmit={(payload) => {
              const finalPayload = {
                firebase_id: selectedStudent.id,
                ...payload
              };
              console.log("Submitting business details payload:", finalPayload);
              const token = localStorage.getItem('authToken');
              fetch(`${backendUrl}/api/tutor/edit-business-details`, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(finalPayload)
              })
              .then((res) => {
                if (!res.ok) {
                  alert("Failed to update business details.");
                  throw new Error("Failed to update business details.");
                } else {
                  alert("Business details updated successfully.");
                  return fetch(
                    `${backendUrl}/api/tutor/students/${selectedStudent.id}?tutorUserID=${encodeURIComponent(tutorId)}&tutorEmail=${encodeURIComponent(tutorEmail)}`,
                    {
                      method: 'GET',
                      headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}`,
                      },
                    }
                  );
                }
              })
              .then((res2) => {
                if (!res2.ok) {
                  throw new Error("Failed to refresh student details.");
                }
                return res2.json();
              })
              .then((updatedStudent) => {
                setSelectedStudent(updatedStudent);
              })
              .catch((error) => {
                console.error("Error editing business details:", error);
                alert("Error editing business details.");
              })
              .finally(() => {
                setShowEditBusinessDialog(false);
              });
            }}
            tutorFullName={tutorName}
          />
        )}
      </Box>
    );

    return enableSearch ? (
      <StyledExpandedStudentView sx={{ padding: 3, margin: 2 }}>
        <Button variant="outlined" onClick={() => setSelectedStudent(null)} sx={{ marginBottom: 2 }}>
          Back to Students List
        </Button>
        {ExpandedViewContent}
      </StyledExpandedStudentView>
    ) : (
      <Box sx={{ padding: 2 }}>
        <Button variant="outlined" onClick={() => setSelectedStudent(null)} sx={{ marginBottom: 2 }}>
          Back to Students List
        </Button>
        {ExpandedViewContent}
      </Box>
    );
  }

  return (
    <Box>
      {enableSearch && (
        <Box sx={{ marginBottom: 2 }}>
          <Autocomplete
            options={studentDetails}
            getOptionLabel={(option) => option.personal?.name || ''}
            renderInput={(params) => <TextField {...params} label="Search By Name" variant="outlined" />}
            onChange={(event, value) => {
              if (value) {
                setSelectedStudent(value);
              }
            }}
          />
        </Box>
      )}
      <Grid container spacing={2}>
        {currentStudents.map((student) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={student.id}>
            {enableSearch ? (
              <StyledStudentCard
                onClick={() => setSelectedStudent(student)}
                sx={{
                  cursor: 'pointer'
                }}
              >
                <CardContent>
                  <Typography variant="h6" align="center">
                    {student.personal?.name || 'Unknown'}
                  </Typography>
                  <Typography variant="body2" align="center" color="textSecondary">
                    {student.business?.test_focus || 'No Focus'}
                  </Typography>
                </CardContent>
              </StyledStudentCard>
            ) : (
              <Card
                variant="outlined"
                onClick={() => setSelectedStudent(student)}
                sx={{
                  minHeight: '120px',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'center',
                  alignItems: 'center',
                  padding: 1,
                  cursor: 'pointer',
                  '&:hover': { boxShadow: '0 4px 12px rgba(0,0,0,0.2)' }
                }}
              >
                <CardContent>
                  <Typography variant="h6" align="center">
                    {student.personal?.name || 'Unknown'}
                  </Typography>
                  <Typography variant="body2" align="center" color="textSecondary">
                    {student.business?.test_focus || 'No Focus'}
                  </Typography>
                </CardContent>
              </Card>
            )}
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
  const [todayStudentNames, setTodayStudentNames] = useState([]);
  const [todaysStudentDetails, setTodaysStudentDetails] = useState([]);
  const [selectedTodayStudent, setSelectedTodayStudent] = useState(null);
  const [todayLoading, setTodayLoading] = useState(false);
  const [calendarLoaded, setCalendarLoaded] = useState(false);
  const backendUrl = process.env.REACT_APP_API_BASE_URL;
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

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

  useEffect(() => {
    async function fetchTodaysStudentsDetail() {
      if (!todayStudentNames || todayStudentNames.length === 0) {
        setTodaysStudentDetails([]);
        return;
      }
      if (!calendarLoaded) return;
      setTodayLoading(true);
      try {
        const token = localStorage.getItem('authToken');
        const res = await fetch(`${backendUrl}/api/tutor/fetch-students-by-names`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({ names: todayStudentNames }),
        });
        if (!res.ok) {
          console.error('Failed to fetch students by name. Status:', res.status);
          setTodayLoading(false);
          return;
        }
        const fetchedDetails = await res.json();
        setTodaysStudentDetails(fetchedDetails);
      } catch (error) {
        console.error('Error fetching todays students by name:', error);
      } finally {
        setTodayLoading(false);
      }
    }
    fetchTodaysStudentsDetail();
  }, [todayStudentNames, backendUrl, calendarLoaded]);

  const combinedTodayStudents = todayStudentNames.map((rawName) => {
    const match = todaysStudentDetails.find(
      (st) =>
        st.personal?.name?.trim().toLowerCase() === rawName.trim().toLowerCase()
    );
    if (match) {
      return { found: true, data: match };
    }
    return {
      found: false,
      data: {
        id: 'NOT_FOUND_' + rawName,
        personal: { name: rawName },
        business: {},
      },
    };
  });

  const handleSelectTodayStudent = (studentObj) => {
    setSelectedTodayStudent(studentObj);
  };
  const handleDeselectTodayStudent = () => {
    setSelectedTodayStudent(null);
  };

  function sortSubcollections(st) {
    if (!st) return st;
    const clone = { ...st };
    const { homeworkCompletion = [], testData = [], testDates = [], goals = [] } = clone;
    clone.homeworkCompletion =
      homeworkCompletion.slice().sort((a, b) => {
        if (a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
        if (a.date && b.date) return new Date(b.date) - new Date(a.date);
        return 0;
      });
    clone.testData =
      testData.slice().sort((a, b) => {
        if (a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
        if (a.date && b.date) return new Date(b.date) - new Date(a.date);
        return 0;
      });
    clone.testDates =
      testDates.slice().sort((a, b) => {
        if (a.test_date && b.test_date) return new Date(b.test_date) - new Date(a.test_date);
        return 0;
      });
    clone.goals =
      goals.slice().sort((a, b) => {
        if (a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
        if (a.date && b.date) return new Date(b.date) - new Date(a.date);
        return 0;
      });
    return clone;
  }

  const [innerTab, setInnerTab] = useState(0);
  const handleInnerTabChange = (event, newValue) => {
    setInnerTab(newValue);
  };

  const handleSignOut = () => {
    localStorage.removeItem('authToken');
    updateToken(null);
    navigate('/');
  };

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  const handleEditHomework = (item) => {
    console.log("Editing homework item in TutorDashboard:", item);
  };

  const handleDeleteHomeworkCompletionToday = async (item) => {
    const confirmDelete = window.confirm("Are you sure you want to delete this homework completion record?");
    if (!confirmDelete) return;
    try {
      const token = localStorage.getItem('authToken');
      const res = await fetch(`${backendUrl}/api/tutor/delete-homework-completion`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          firebase_id: selectedTodayStudent.id,
          event_id: item.id,
        }),
      });
      if (!res.ok) {
        console.error('Failed to delete homework completion. Status:', res.status);
        alert("Failed to delete homework completion record.");
        return;
      }
      alert("Homework completion record deleted successfully.");
      const res2 = await fetch(
        `${backendUrl}/api/tutor/students/${selectedTodayStudent.id}?tutorUserID=${encodeURIComponent(profile.user_id)}&tutorEmail=${encodeURIComponent(profile.email)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        }
      );
      if (res2.ok) {
        const updatedStudent = await res2.json();
        setSelectedTodayStudent(updatedStudent);
      }
    } catch (error) {
      console.error('Error deleting homework completion record:', error);
      alert("Error deleting homework completion record.");
    }
  };

  const handleDeleteEvent = async (item) => {
    const confirmDelete = window.confirm("Are you sure you want to delete this event?");
    if (!confirmDelete) return;
    try {
      const token = localStorage.getItem('authToken');
      const res = await fetch(`${backendUrl}/api/tutor/delete-event`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          firebase_id: selectedTodayStudent.id,
          event_id: item.id,
        }),
      });
      if (!res.ok) {
        console.error('Failed to delete event. Status:', res.status);
        alert("Failed to delete event.");
        return;
      }
      alert("Event deleted successfully.");
      const res2 = await fetch(
        `${backendUrl}/api/tutor/students/${selectedTodayStudent.id}?tutorUserID=${encodeURIComponent(profile.user_id)}&tutorEmail=${encodeURIComponent(profile.email)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        }
      );
      if (res2.ok) {
        const updatedStudent = await res2.json();
        setSelectedTodayStudent(updatedStudent);
      }
    } catch (error) {
      console.error('Error deleting event:', error);
      alert("Error deleting event.");
    }
  };

  return (
    <RootContainer>
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
      <Container maxWidth="xl" sx={{ marginTop: '24px' }}>
        <HeroSection>
          <Typography variant="h4" sx={{ fontWeight: 700 }}>
            Your Schedule Today:
          </Typography>
          {profile && profile.user_id ? (
            <Box sx={{ marginTop: '16px', paddingRight: '40px' }}>
              <TodaySchedule
                tutorId={profile.user_id}
                backendUrl={backendUrl}
                onCalendarLoaded={() => setCalendarLoaded(true)}
                onStudentNamesUpdate={(names) => {
                  const oldSet = new Set(todayStudentNames.map((n) => n.trim()));
                  const newSet = new Set(names.map((n) => n.trim()));
                  if (oldSet.size === newSet.size) {
                    let allSame = true;
                    for (let n of newSet) {
                      if (!oldSet.has(n)) {
                        allSame = false;
                        break;
                      }
                    }
                    if (allSame) return;
                  }
                  setTodayStudentNames(names);
                }}
              />
            </Box>
          ) : (
            <Typography variant="body1" sx={{ opacity: 0.9, marginTop: '8px' }}>
              Loading your schedule...
            </Typography>
          )}
        </HeroSection>
      </Container>
      {profile && profile.user_id && profile.email && (
        <Container maxWidth="xl" sx={{ marginBottom: '24px' }}>
          <Typography variant="h4" sx={{ fontWeight: 700, marginBottom: '16px' }}>
            Today's Appointments:
          </Typography>
          <StudentsTab
            tutorId={profile.user_id}
            tutorEmail={profile.email}
            backendUrl={backendUrl}
            filterTodayAppointments={true}
            tutorName={profile.name}
          />
        </Container>
      )}
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
              <Tab label="Today's Students" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="My Students" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="My Schedule" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
              <Tab label="Tutor Tools" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
            </Tabs>
          </Box>
          <TabPanel value={activeTab} index={0}>
            <SectionContainer>
              <SectionTitle variant="h6">Today's Students</SectionTitle>
              {selectedTodayStudent ? (
                <Box sx={{ padding: 2 }}>
                  <Button variant="outlined" onClick={() => setSelectedTodayStudent(null)} sx={{ marginBottom: 2 }}>
                    Back to Students List
                  </Button>
                  <Typography variant="h5" sx={{ marginBottom: 2 }}>
                    {selectedTodayStudent.personal?.name || 'Student Overview'}
                  </Typography>
                  <Box sx={{ marginBottom: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center' }}>
                      <Typography variant="subtitle1">Personal Details:</Typography>
                      <IconButton size="small" onClick={() => console.log("Edit Personal Details", selectedTodayStudent)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Box>
                    {Object.entries(selectedTodayStudent.personal || {}).map(([key, value]) => (
                      <Typography key={key}>
                        <strong>{key}:</strong> {value}
                      </Typography>
                    ))}
                  </Box>
                  <Box sx={{ marginBottom: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center' }}>
                      <Typography variant="subtitle1">Business Details:</Typography>
                      <IconButton size="small" onClick={() => console.log("Edit Business Details", selectedTodayStudent)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Box>
                    <Typography>
                      <strong>Team Lead:</strong> {selectedTodayStudent.business?.team_lead || 'N/A'}
                    </Typography>
                    <Typography>
                      <strong>Test Focus:</strong> {selectedTodayStudent.business?.test_focus || 'N/A'}
                    </Typography>
                  </Box>
                  <Tabs value={innerTab} onChange={(e, newValue) => setInnerTab(newValue)}>
                    <Tab label="Homework Completion" />
                    <Tab label="Test Data" />
                    <Tab label="Test Dates" />
                    <Tab label="Goals" />
                    <Tab label="Events" />
                  </Tabs>
                  <TabPanel value={innerTab} index={0}>
                    {selectedTodayStudent.homeworkCompletion && selectedTodayStudent.homeworkCompletion.length > 0 ? (
                      selectedTodayStudent.homeworkCompletion.slice().sort((a, b) => {
                        if(a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
                        if(a.date && b.date) return new Date(b.date) - new Date(a.date);
                        return 0;
                      }).map((item) => (
                        <InfoCard key={item.id} item={item} onEdit={handleEditHomework} onDelete={handleDeleteHomeworkCompletionToday} />
                      ))
                    ) : (
                      <Typography>No Homework Completion data available.</Typography>
                    )}
                  </TabPanel>
                  <TabPanel value={innerTab} index={1}>
                    {selectedTodayStudent.testData && selectedTodayStudent.testData.length > 0 ? (
                      selectedTodayStudent.testData.slice().sort((a, b) => {
                        if(a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
                        if(a.date && b.date) return new Date(b.date) - new Date(a.date);
                        return 0;
                      }).map((item) => <InfoCard key={item.id} item={item} />)
                    ) : (
                      <Typography>No Test Data available.</Typography>
                    )}
                  </TabPanel>
                  <TabPanel value={innerTab} index={2}>
                    {selectedTodayStudent.testDates && selectedTodayStudent.testDates.length > 0 ? (
                      selectedTodayStudent.testDates.slice().sort((a, b) => {
                        if(a.test_date && b.test_date) return new Date(b.test_date) - new Date(a.test_date);
                        return 0;
                      }).map((item) => <InfoCard key={item.id} item={item} />)
                    ) : (
                      <Typography>No Test Dates available.</Typography>
                    )}
                  </TabPanel>
                  <TabPanel value={innerTab} index={3}>
                    {selectedTodayStudent.goals && selectedTodayStudent.goals.length > 0 ? (
                      selectedTodayStudent.goals.slice().sort((a, b) => {
                        if(a.timestamp && b.timestamp) return b.timestamp - a.timestamp;
                        if(a.date && b.date) return new Date(b.date) - new Date(a.date);
                        return 0;
                      }).map((item) => <InfoCard key={item.id} item={item} />)
                    ) : (
                      <Typography>No Goals available.</Typography>
                    )}
                  </TabPanel>
                  <TabPanel value={innerTab} index={4}>
                    {selectedTodayStudent.events && selectedTodayStudent.events.length > 0 ? (
                      selectedTodayStudent.events.map((item) => (
                        <InfoCard key={item.id} item={item} onDelete={handleDeleteEvent} />
                      ))
                    ) : (
                      <Typography>No Events available.</Typography>
                    )}
                  </TabPanel>
                </Box>
              ) : (
                <>
                  {combinedTodayStudents.length > 0 ? (
                    <Grid container spacing={2}>
                      {combinedTodayStudents.map((studentObj, index) => (
                        <Grid item xs={12} sm={6} md={4} lg={3} key={index}>
                          <Card
                            variant="outlined"
                            onClick={() => {
                              if (studentObj.found) {
                                setSelectedTodayStudent(studentObj.data);
                              }
                            }}
                            sx={{
                              minHeight: '120px',
                              display: 'flex',
                              flexDirection: 'column',
                              justifyContent: 'center',
                              alignItems: 'center',
                              padding: 1,
                              cursor: studentObj.found ? 'pointer' : 'default',
                              '&:hover': studentObj.found ? { boxShadow: '0 4px 12px rgba(0,0,0,0.2)' } : {}
                            }}
                          >
                            <CardContent>
                              <Typography variant="h6" align="center">
                                {studentObj.data.personal?.name || 'Unknown'}
                              </Typography>
                              {!studentObj.found && (
                                <Typography variant="body2" align="center" color="error">
                                  Not Associated
                                </Typography>
                              )}
                            </CardContent>
                          </Card>
                        </Grid>
                      ))}
                    </Grid>
                  ) : (
                    <Typography variant="body1">No students available.</Typography>
                  )}
                </>
              )}
            </SectionContainer>
          </TabPanel>
          <TabPanel value={activeTab} index={1}>
            <SectionContainer>
              {profile && profile.user_id && profile.email ? (
                associationComplete ? (
                  <StudentsTab
                    tutorId={profile.user_id}
                    tutorEmail={profile.email}
                    backendUrl={backendUrl}
                    enableSearch={true}
                    tutorName={profile.name}
                  />
                ) : (
                  <Typography variant="body1">Updating your student associations...</Typography>
                )
              ) : (
                <Typography variant="body1">Loading your students...</Typography>
              )}
            </SectionContainer>
          </TabPanel>
          <TabPanel value={activeTab} index={2}>
            <SectionContainer>
              <SectionTitle variant="h6">My Schedule</SectionTitle>
              {profile && profile.user_id ? (
                <MySchedule tutorId={profile.user_id} backendUrl={backendUrl} />
              ) : (
                <Typography variant="body1">Loading your schedule...</Typography>
              )}
            </SectionContainer>
          </TabPanel>
          <TabPanel value={activeTab} index={3}>
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
