// StudentIntakeDialog.jsx
import React, { useState, useEffect } from 'react';
import {
  Dialog, DialogTitle, DialogContent, DialogActions,
  Button, TextField, FormControl, InputLabel, Select, MenuItem,
  Checkbox, FormControlLabel, Box, Typography
} from '@mui/material';

const schedulerOptions = ['Parents', 'Students', 'TBD'];
const gradeOptions = ['9', '10', '11', '12', 'other'];
const testFocusOptions = ['SAT', 'ACT', 'PSAT', 'PACT', 'TBD'];
const accommodationsOptions = ['None', '1.5x', '2x', 'Multi Day Testing', 'Other', 'Unknown'];

function StudentIntakeDialog({ open, onClose, onSubmit }) {
  const [lastName, setLastName] = useState('');
  const [firstName, setFirstName] = useState('');
  const [studentEmail, setStudentEmail] = useState('');
  const [studentNumber, setStudentNumber] = useState('');
  const [parentEmail, setParentEmail] = useState('');
  const [parentNumber, setParentNumber] = useState('');
  const [scheduler, setScheduler] = useState('');
  const [school, setSchool] = useState('');
  const [grade, setGrade] = useState('');
  const [otherGrade, setOtherGrade] = useState('');
  const [testFocus, setTestFocus] = useState('');
  const [accommodations, setAccommodations] = useState('');
  const [interests, setInterests] = useState('');
  const [availability, setAvailability] = useState('');
  const [registeredForTest, setRegisteredForTest] = useState(false);
  const [testDate, setTestDate] = useState('');
  const [errors, setErrors] = useState({});

  useEffect(() => {
    if (open) {
      setLastName('');
      setFirstName('');
      setStudentEmail('');
      setStudentNumber('');
      setParentEmail('');
      setParentNumber('');
      setScheduler('');
      setSchool('');
      setGrade('');
      setOtherGrade('');
      setTestFocus('');
      setAccommodations('');
      setInterests('');
      setAvailability('');
      setRegisteredForTest(false);
      setTestDate('');
      setErrors({});
    }
  }, [open]);

  const validate = () => {
    const newErrors = {};
    if (!lastName) newErrors.lastName = 'Required';
    if (!firstName) newErrors.firstName = 'Required';
    if (!studentEmail) newErrors.studentEmail = 'Required';
    else if (!/\S+@\S+\.\S+/.test(studentEmail)) newErrors.studentEmail = 'Invalid email';
    if (!parentEmail) newErrors.parentEmail = 'Required';
    else if (!/\S+@\S+\.\S+/.test(parentEmail)) newErrors.parentEmail = 'Invalid email';
    if (!scheduler) newErrors.scheduler = 'Required';
    if (!grade) newErrors.grade = 'Required';
    else if (grade === 'other' && (!otherGrade || isNaN(Number(otherGrade)))) newErrors.otherGrade = 'Enter a valid number';
    if (!testFocus) newErrors.testFocus = 'Required';
    if (!accommodations) newErrors.accommodations = 'Required';
    if (!interests) newErrors.interests = 'Required';
    if (!availability) newErrors.availability = 'Required';
    if (registeredForTest && !testDate) newErrors.testDate = 'Required';
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;
    onSubmit({
      lastName,
      firstName,
      studentEmail,
      studentNumber,
      parentEmail,
      parentNumber,
      scheduler,
      school,
      grade: grade === 'other' ? otherGrade : grade,
      testFocus,
      accommodations,
      interests,
      availability,
      registeredForTest,
      testDate,
    });
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Student Intake</DialogTitle>
      <DialogContent>
        <Box component="form" noValidate sx={{ mt: 2 }}>
          <TextField
            label="Student Last Name"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            fullWidth
            required
            margin="normal"
            error={!!errors.lastName}
            helperText={errors.lastName}
          />
          <TextField
            label="Student First Name"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            fullWidth
            required
            margin="normal"
            error={!!errors.firstName}
            helperText={errors.firstName}
          />
          <TextField
            label="Student Email"
            type="email"
            value={studentEmail}
            onChange={(e) => setStudentEmail(e.target.value)}
            fullWidth
            required
            margin="normal"
            error={!!errors.studentEmail}
            helperText={errors.studentEmail}
          />
          <TextField
            label="Student Number"
            value={studentNumber}
            onChange={(e) => setStudentNumber(e.target.value)}
            fullWidth
            margin="normal"
          />
          <TextField
            label="Parent Email"
            type="email"
            value={parentEmail}
            onChange={(e) => setParentEmail(e.target.value)}
            fullWidth
            required
            margin="normal"
            error={!!errors.parentEmail}
            helperText={errors.parentEmail}
          />
          <TextField
            label="Parent Number"
            value={parentNumber}
            onChange={(e) => setParentNumber(e.target.value)}
            fullWidth
            margin="normal"
          />
          <FormControl fullWidth margin="normal" required error={!!errors.scheduler}>
            <InputLabel>Who Schedules Meetings?</InputLabel>
            <Select
              value={scheduler}
              label="Who Schedules Meetings?"
              onChange={(e) => setScheduler(e.target.value)}
            >
              {schedulerOptions.map((opt) => (
                <MenuItem key={opt} value={opt}>{opt}</MenuItem>
              ))}
            </Select>
            <Typography variant="caption" color="error">{errors.scheduler}</Typography>
          </FormControl>
          <TextField
            label="School"
            value={school}
            onChange={(e) => setSchool(e.target.value)}
            fullWidth
            margin="normal"
          />
          <FormControl fullWidth margin="normal" required error={!!errors.grade}>
            <InputLabel>Grade</InputLabel>
            <Select
              value={grade}
              label="Grade"
              onChange={(e) => setGrade(e.target.value)}
            >
              {gradeOptions.map((opt) => (
                <MenuItem key={opt} value={opt}>{opt}</MenuItem>
              ))}
            </Select>
            {grade === 'other' && (
              <TextField
                label="Specify Grade"
                value={otherGrade}
                onChange={(e) => setOtherGrade(e.target.value)}
                fullWidth
                margin="normal"
                error={!!errors.otherGrade}
                helperText={errors.otherGrade}
              />
            )}
            <Typography variant="caption" color="error">{errors.grade}</Typography>
          </FormControl>
          <FormControl fullWidth margin="normal" required error={!!errors.testFocus}>
            <InputLabel>Test Focus</InputLabel>
            <Select
              value={testFocus}
              label="Test Focus"
              onChange={(e) => setTestFocus(e.target.value)}
            >
              {testFocusOptions.map((opt) => (
                <MenuItem key={opt} value={opt}>{opt}</MenuItem>
              ))}
            </Select>
            <Typography variant="caption" color="error">{errors.testFocus}</Typography>
          </FormControl>
          <FormControl fullWidth margin="normal" required error={!!errors.accommodations}>
            <InputLabel>Accommodations (504 Plan)</InputLabel>
            <Select
              value={accommodations}
              label="Accommodations (504 Plan)"
              onChange={(e) => setAccommodations(e.target.value)}
            >
              {accommodationsOptions.map((opt) => (
                <MenuItem key={opt} value={opt}>{opt}</MenuItem>
              ))}
            </Select>
            <Typography variant="caption" color="error">{errors.accommodations}</Typography>
          </FormControl>
          <TextField
            label="Interests"
            value={interests}
            onChange={(e) => setInterests(e.target.value)}
            fullWidth
            required
            margin="normal"
            error={!!errors.interests}
            helperText={errors.interests}
          />
          <TextField
            label="Availability"
            value={availability}
            onChange={(e) => setAvailability(e.target.value)}
            fullWidth
            required
            margin="normal"
            error={!!errors.availability}
            helperText={errors.availability}
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={registeredForTest}
                onChange={(e) => setRegisteredForTest(e.target.checked)}
              />
            }
            label="Registered for Upcoming Tests"
          />
          {registeredForTest && (
            <TextField
              label="Test Date"
              type="date"
              value={testDate}
              onChange={(e) => setTestDate(e.target.value)}
              fullWidth
              margin="normal"
              required
              InputLabelProps={{ shrink: true }}
              error={!!errors.testDate}
              helperText={errors.testDate}
            />
          )}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit} variant="contained">Submit</Button>
      </DialogActions>
    </Dialog>
  );
}

export default StudentIntakeDialog;