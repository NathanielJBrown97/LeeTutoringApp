import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Box,
  Typography,
  Slider,
  Grid
} from '@mui/material';

const attendanceOptions = [
  "On Time",
  "Late",
  "Ended Early",
  "No Show"
];

const durationOptions = [
  { value: "1", label: "1:00" },
  { value: "0.25", label: "0:15" },
  { value: "0.5", label: "0:30" },
  { value: "0.75", label: "0:45" },
  { value: "1.25", label: "1:15" },
  { value: "1.5", label: "1:30" },
  { value: "1.75", label: "1:45" },
  { value: "2", label: "2:00" }
];

// Converts a date in "MM/DD/YYYY" format (from DB) to "YYYY-MM-DD" format (for the input)
const convertDateForInput = (dateWithSlashes) => {
  if (!dateWithSlashes) return "";
  const parts = dateWithSlashes.split("/");
  if (parts.length !== 3) return "";
  return `${parts[2]}-${parts[0].padStart(2, '0')}-${parts[1].padStart(2, '0')}`;
};

// Converts a date in "YYYY-MM-DD" format (from input) to "MM/DD/YYYY" format (for the backend)
const formatDateWithSlashes = (dateStr) => {
  if (!dateStr) return '';
  const parts = dateStr.split('-');
  if (parts.length !== 3) return '';
  return `${parts[1].padStart(2, '0')}/${parts[2].padStart(2, '0')}/${parts[0]}`;
};

const buildDateId = (dateStr) => {
  if (!dateStr) return '';
  const parts = dateStr.split('-');
  if (parts.length !== 3) return '';
  return `${parts[1].padStart(2, '0')}-${parts[2].padStart(2, '0')}-${parts[0]}`;
};

const EditHomeworkCompletionDialog = ({ open, onClose, onSubmit, initialData, tutorFullName }) => {
  // initialData is expected to have keys:
  // date (in "MM/DD/YYYY"), attendance, duration, feedback, engagement, percentage_complete, and timestamp.
  // We need to convert the date to the "YYYY-MM-DD" format for the <TextField type="date">.
  const [date, setDate] = useState('');
  const [attendance, setAttendance] = useState('');
  const [duration, setDuration] = useState('');
  const [feedback, setFeedback] = useState('');
  const [engagementLevel, setEngagementLevel] = useState(0);
  const [percentageComplete, setPercentageComplete] = useState(0);

  useEffect(() => {
    if (open && initialData) {
      setDate(convertDateForInput(initialData.date)); // convert from "MM/DD/YYYY" to "YYYY-MM-DD"
      setAttendance(initialData.attendance || '');
      setDuration(initialData.duration || '');
      setFeedback(initialData.feedback || '');
      setEngagementLevel(parseFloat(initialData.engagement) || 0);
      setPercentageComplete(parseFloat(initialData.percentage_complete) || 0);
    }
  }, [open, initialData]);

  const handleSubmit = () => {
    if (!date || !attendance || !duration) {
      alert('Please fill out the date, attendance, and duration fields.');
      return;
    }
    // Convert the date from "YYYY-MM-DD" to "MM/DD/YYYY" for storage
    const formattedDate = formatDateWithSlashes(date);
    // Extract only the tutor's first name.
    const tutorFirstName = tutorFullName ? tutorFullName.split(' ')[0] : '';
    // Generate a new timestamp for the update.
    const timestamp = new Date().toISOString();

    const payload = {
      date: formattedDate, // stored with slashes, e.g. "02/26/2025"
      attendance,          // string e.g., "On Time"
      duration,            // string e.g., "0.25"
      feedback,
      percentage_complete: percentageComplete.toString(),
      engagement: engagementLevel.toString(),
      tutor: tutorFirstName,
      timestamp            // ISO format e.g. "2025-02-11T00:40:05Z"
    };

    onSubmit(payload);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Homework Completion</DialogTitle>
      <DialogContent>
        <Box component="form" noValidate sx={{ mt: 2 }}>
          {/* Date Picker */}
          <TextField
            label="Select Date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            fullWidth
            required
            InputLabelProps={{ shrink: true }}
            margin="normal"
          />
          {/* Attendance Dropdown */}
          <FormControl fullWidth margin="normal" required>
            <InputLabel id="attendance-select-label">Attendance</InputLabel>
            <Select
              labelId="attendance-select-label"
              value={attendance}
              label="Attendance"
              onChange={(e) => setAttendance(e.target.value)}
            >
              {attendanceOptions.map((option) => (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          {/* Duration Dropdown */}
          <FormControl fullWidth margin="normal" required>
            <InputLabel id="duration-select-label">Duration</InputLabel>
            <Select
              labelId="duration-select-label"
              value={duration}
              label="Duration"
              onChange={(e) => setDuration(e.target.value)}
            >
              {durationOptions.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          {/* Feedback Input */}
          <TextField
            label="Feedback"
            type="text"
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            fullWidth
            margin="normal"
          />
          {/* Engagement Level Slider */}
          <Box sx={{ mt: 3 }}>
            <Typography gutterBottom>Engagement Level ({engagementLevel}%)</Typography>
            <Slider
              value={engagementLevel}
              onChange={(e, newValue) => setEngagementLevel(newValue)}
              aria-labelledby="engagement-level-slider"
              valueLabelDisplay="auto"
              step={1}
              min={0}
              max={100}
            />
          </Box>
          {/* Percentage Complete Slider */}
          <Box sx={{ mt: 3 }}>
            <Typography gutterBottom>Percentage Complete ({percentageComplete}%)</Typography>
            <Slider
              value={percentageComplete}
              onChange={(e, newValue) => setPercentageComplete(newValue)}
              aria-labelledby="percentage-complete-slider"
              valueLabelDisplay="auto"
              step={1}
              min={0}
              max={100}
            />
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit} variant="contained">
          Save Changes
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default EditHomeworkCompletionDialog;
