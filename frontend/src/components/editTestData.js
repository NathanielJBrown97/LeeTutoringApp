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
  FormControlLabel,
  Radio,
  RadioGroup,
  Box,
  Grid,
  Typography,
} from '@mui/material';

// Conversion functions must be declared before they are used.
const convertDateForInput = (dateStr) => {
  if (!dateStr) return '';
  const parts = dateStr.split('/');
  if (parts.length !== 3) return dateStr;
  const [month, day, year] = parts;
  return `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
};

const formatDate = (dateStr) => {
  if (!dateStr) return '';
  const [year, month, day] = dateStr.split('-');
  return `${parseInt(month)}/${parseInt(day)}/${year}`;
};

const EditTestDataDialog = ({ open, onClose, onSubmit, initialData }) => {
  // Initialize state with initialData if available; otherwise, empty strings.
  const [testDate, setTestDate] = useState(initialData ? convertDateForInput(initialData.date) : '');
  const [baseline, setBaseline] = useState(initialData ? String(initialData.baseline) : '');
  const [test, setTest] = useState(initialData?.test || '');
  const [type, setType] = useState(initialData?.type || '');
  const [actScores, setActScores] = useState({
    ACT_Total: initialData?.ACT_Scores?.ACT_Total || '',
    English: initialData?.ACT_Scores?.English || '',
    Math: initialData?.ACT_Scores?.Math || '',
    Reading: initialData?.ACT_Scores?.Reading || '',
    Science: initialData?.ACT_Scores?.Science || ''
  });
  const [satScores, setSatScores] = useState({
    SAT_Total: initialData?.SAT_Scores?.SAT_Total || '',
    EBRW: initialData?.SAT_Scores?.EBRW || '',
    Math: initialData?.SAT_Scores?.Math || '',
    Reading: initialData?.SAT_Scores?.Reading || '',
    Writing: initialData?.SAT_Scores?.Writing || ''
  });

  // When the dialog opens or initialData changes, prefill the form.
  useEffect(() => {
    if (open && initialData) {
      setTestDate(convertDateForInput(initialData.date));
      setBaseline(String(initialData.baseline));
      setTest(initialData.test || '');
      setType(initialData.type || '');
      setActScores({
        ACT_Total: initialData?.ACT_Scores?.ACT_Total || '',
        English: initialData?.ACT_Scores?.English || '',
        Math: initialData?.ACT_Scores?.Math || '',
        Reading: initialData?.ACT_Scores?.Reading || '',
        Science: initialData?.ACT_Scores?.Science || ''
      });
      setSatScores({
        SAT_Total: initialData?.SAT_Scores?.SAT_Total || '',
        EBRW: initialData?.SAT_Scores?.EBRW || '',
        Math: initialData?.SAT_Scores?.Math || '',
        Reading: initialData?.SAT_Scores?.Reading || '',
        Writing: initialData?.SAT_Scores?.Writing || ''
      });
    }
  }, [open, initialData]);

  const buildScoreObject = (scores) => {
    const result = {};
    Object.keys(scores).forEach((key) => {
      if (scores[key] !== '') {
        result[key] = Number(scores[key]);
      }
    });
    return result;
  };

  const handleSubmit = () => {
    if (!testDate || baseline === '' || !test || !type) {
      alert('Please fill out all required fields.');
      return;
    }
    if ((test === 'ACT' || test === 'PACT') && !actScores.ACT_Total) {
      alert('Please fill out at least the ACT_Total score for ACT or PACT tests.');
      return;
    }
    if ((test === 'SAT' || test === 'PSAT') && !satScores.SAT_Total) {
      alert('Please fill out at least the SAT_Total score for SAT or PSAT tests.');
      return;
    }

    const formattedDate = formatDate(testDate);
    const idDate = formattedDate.replace(/\//g, '-');
    const id = `${type} ${test} ${idDate}`;

    const payload = {
      date: formattedDate,
      baseline: baseline === 'true',
      test,
      type,
      id,
      ACT_Scores: buildScoreObject(actScores),
      SAT_Scores: buildScoreObject(satScores)
    };

    onSubmit(payload);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Test Data</DialogTitle>
      <DialogContent>
        <Box component="form" noValidate sx={{ mt: 2 }}>
          <TextField
            label="Select Date of Test"
            type="date"
            value={testDate}
            onChange={(e) => setTestDate(e.target.value)}
            fullWidth
            required
            InputLabelProps={{ shrink: true }}
            margin="normal"
          />
          <FormControl component="fieldset" margin="normal" required>
            <Typography variant="subtitle1">Baseline</Typography>
            <RadioGroup row value={baseline} onChange={(e) => setBaseline(e.target.value)}>
              <FormControlLabel value="true" control={<Radio />} label="True" />
              <FormControlLabel value="false" control={<Radio />} label="False" />
            </RadioGroup>
          </FormControl>
          <FormControl fullWidth margin="normal" required>
            <InputLabel id="test-select-label">Test</InputLabel>
            <Select
              labelId="test-select-label"
              value={test}
              label="Test"
              onChange={(e) => setTest(e.target.value)}
            >
              <MenuItem value="ACT">ACT</MenuItem>
              <MenuItem value="SAT">SAT</MenuItem>
              <MenuItem value="PSAT">PSAT</MenuItem>
              <MenuItem value="PACT">PACT</MenuItem>
            </Select>
          </FormControl>
          <FormControl fullWidth margin="normal" required>
            <InputLabel id="type-select-label">Type</InputLabel>
            <Select
              labelId="type-select-label"
              value={type}
              label="Type"
              onChange={(e) => setType(e.target.value)}
            >
              <MenuItem value="Official">Official</MenuItem>
              <MenuItem value="Unofficial">Unofficial</MenuItem>
              <MenuItem value="SS Official">SS Official</MenuItem>
              <MenuItem value="SS Unofficial">SS Unofficial</MenuItem>
            </Select>
          </FormControl>
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6">ACT Scores (Optional)</Typography>
            <Grid container spacing={2}>
              <Grid item xs={6}>
                <TextField
                  label="ACT_Total"
                  type="number"
                  value={actScores.ACT_Total}
                  onChange={(e) => setActScores({ ...actScores, ACT_Total: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="English"
                  type="number"
                  value={actScores.English}
                  onChange={(e) => setActScores({ ...actScores, English: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="Math"
                  type="number"
                  value={actScores.Math}
                  onChange={(e) => setActScores({ ...actScores, Math: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="Reading"
                  type="number"
                  value={actScores.Reading}
                  onChange={(e) => setActScores({ ...actScores, Reading: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="Science"
                  type="number"
                  value={actScores.Science}
                  onChange={(e) => setActScores({ ...actScores, Science: e.target.value })}
                  fullWidth
                />
              </Grid>
            </Grid>
          </Box>
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6">SAT Scores (Optional)</Typography>
            <Grid container spacing={2}>
              <Grid item xs={6}>
                <TextField
                  label="SAT_Total"
                  type="number"
                  value={satScores.SAT_Total}
                  onChange={(e) => setSatScores({ ...satScores, SAT_Total: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="EBRW"
                  type="number"
                  value={satScores.EBRW}
                  onChange={(e) => setSatScores({ ...satScores, EBRW: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="Math"
                  type="number"
                  value={satScores.Math}
                  onChange={(e) => setSatScores({ ...satScores, Math: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="Reading"
                  type="number"
                  value={satScores.Reading}
                  onChange={(e) => setSatScores({ ...satScores, Reading: e.target.value })}
                  fullWidth
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="Writing"
                  type="number"
                  value={satScores.Writing}
                  onChange={(e) => setSatScores({ ...satScores, Writing: e.target.value })}
                  fullWidth
                />
              </Grid>
            </Grid>
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit} variant="contained">
          Submit
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default EditTestDataDialog;
