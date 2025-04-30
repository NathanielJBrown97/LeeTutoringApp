import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Radio,
  RadioGroup,
  FormControlLabel,
  FormControl,
  FormLabel,
  Checkbox,
  TextField,
  InputLabel,
  Select,
  MenuItem,
  Typography
} from '@mui/material';

// === Constants replicating form (test ID) options from your HTML ===
const ACT_FORMS = [
  'F07', 'B05', 'E26', 'E25', 'G01', '67c', 'D03', 'E23', 'F11', 'F12',
  'G19', '16mc1', '65d', '64e', '61c', '59f', '63e', '16mc2', '16mc3',
  '1mc', 'C03', 'D06', '2176', 'H31'
];

const SAT_FORMS = [
  'sat1', 'sat2', 'sat3', 'sat4', 'sat5', 'sat6', 'sat7', 'sat8',
  'sat418', 'sat419', 'sat320', 'sat1020', 'psat1', 'psat2'
];

const DSAT_FORMS = [
  'dsat1', 'dsat2', 'dsat3', 'dsat4', 'dsat5',
  'dsat6', 'dsat7', 'dsat8', 'dsat9', 'dsat10'
];

// DSAT Quiz: separate reading vs math quizzes
const DSAT_QUIZ_READING = [
  'dSAT_Boundaries_3a','dSAT_Inferences_2a','dSAT_Inferences_1a','dSAT_Inferences_1b',
  'dSAT_Inferences_2b','dSAT_Inferences_3a','dSAT_Rhetorical Synthesis_1a',
  'dSAT_Inferences_3b','dSAT_Rhetorical Synthesis_1b','dSAT_Rhetorical Synthesis_2c',
  'dSAT_Rhetorical Synthesis_2b','dSAT_Rhetorical Synthesis_3a','dSAT_Rhetorical Synthesis_2a',
  'dSAT_Boundaries_1a','dSAT_Boundaries_2a','dSAT_Boundaries_1c','dSAT_Boundaries_1b',
  'dSAT_Boundaries_3b','dSAT_Central Ideas_1a','dSAT_Central Ideas_1b','dSAT_Boundaries_2b',
  'dSAT_Central Ideas_3a','dSAT_Central Ideas_2b','dSAT_Central Ideas_3b','dSAT_Command of Evidence_1a',
  'dSAT_Central Ideas_2a','dSAT_Command of Evidence_1c','dSAT_Command of Evidence_2a',
  'dSAT_Command of Evidence_2b','dSAT_Command of Evidence_2c','dSAT_Command of Evidence_1b',
  'dSAT_Command of Evidence_3b','dSAT_Command of Evidence_3a','dSAT_Command of Evidence_3c',
  'dSAT_Cross Text Connections_1a','dSAT_Command of Evidence_2d','dSAT_Cross Text Connections_3b',
  'dSAT_Cross Text Connections_2a','dSAT_Cross Text Connections_3a','dSAT_Form Structure and Sense_1a',
  'dSAT_Cross Text Connections_1b','dSAT_Form Structure and Sense_1b','dSAT_Form Structure and Sense_2b',
  'dSAT_Form Structure and Sense_2a','dSAT_Form Structure and Sense_1d','dSAT_Form Structure and Sense_1c',
  'dSAT_Form Structure and Sense_3b','dSAT_Form Structure and Sense_3a','dSAT_Rhetorical Synthesis_3b',
  'dSAT_Text Structure and Purpose_1b','dSAT_Text Structure and Purpose_2b','dSAT_Text Structure and Purpose_1a',
  'dSAT_Text Structure and Purpose_2a','dSAT_Transitions_1a','dSAT_Transitions_2a','dSAT_Transitions_1b',
  'dSAT_Transitions_1c','dSAT_Transitions_2b','dSAT_Text Structure and Purpose_3a','dSAT_Transitions_3b',
  'dSAT_Words in Context_1b','dSAT_Words in Context_1a','dSAT_Transitions_3a','dSAT_Words in Context_1d',
  'dSAT_Words in Context_2a','dSAT_Words in Context_3a','dSAT_Words in Context_1c'
];

const DSAT_QUIZ_MATH = [
  'dSAT_Nonlinear Functions_2a','dSAT_Nonlinear equations_123a','dSAT_Models and Scatterplots_2a',
  'dSAT_Nonlinear equations in one or two variables_3a','dSAT_Nonlinear equations in one or two variables_2a',
  'dSAT_Nonlinear equations in one or two variables_2b','dSAT_Nonlinear equations in one or two variables_1a',
  'dSAT_Nonlinear equations in one or two variables_3b','dSAT_Models and Scatterplots_3a','dSAT_Nonlinear Functions_1b',
  'dSAT_Nonlinear Functions_1a','dSAT_Nonlinear Functions_3a','dSAT_Nonlinear Functions_3c','dSAT_Nonlinear Functions_3b',
  'dSAT_Nonlinear Functions_2c','dSAT_Nonlinear Functions_2b','dSAT_Percentages_1a','dSAT_Percentages_2a',
  'dSAT_Percentages_1b','dSAT_Percentages_3a','dSAT_Observations and Experiments_123a','dSAT_Ratios rates proportions units_1a',
  'dSAT_Probability_3a','dSAT_Probability_2a','dSAT_Ratios rates proportions units_1b','dSAT_Probability_1a',
  'dSAT_Right triangles and trig_3a','dSAT_Ratios rates proportions units_3a','dSAT_Right triangles and trig_1+2a',
  'dSAT_Ratios rates proportions units_2b','dSAT_Ratios rates proportions units_2a','dSAT_Systems of Linear Equations_1b',
  'dSAT_Systems of Linear Equations_1a','dSAT_Systems of Linear Equations_2a','dSAT_Sample stats and margin of error_3a',
  'dSAT_Sample stats and margin of error_1+2a','dSAT_Systems of Linear Equations_3a','dSAT_Systems of Linear Equations_2b',
  'dSAT_Distributions_1a','dSAT_Distributions_2a','dSAT_Distributions_3a','dSAT_Distributions_1b',
  'dSAT_Circles_1+2a','dSAT_Area and Volume_2a','dSAT_Area and Volume_1a','dSAT_Area and Volume_3a',
  'dSAT_Circles_3a','dSAT_Equivalent Expressions_3a','dSAT_Equivalent Expressions_2a','dSAT_Equivalent Expressions_2b',
  'dSAT_Equivalent Expressions_1b','dSAT_Equivalent Expressions_1a','dSAT_Linear eq two variables_1a',
  'dSAT_Linear eq two variables_1b','dSAT_Linear eq two variables_1c','dSAT_Linear eq two variables_2a',
  'dSAT_Equivalent Expressions_3b','dSAT_Linear Equations One Variable_2a','dSAT_Linear Equations One Variable_3a',
  'dSAT_Linear Equations One Variable_1a','dSAT_Linear Equations One Variable_1b','dSAT_Linear eq two variables_3a',
  'dSAT_Linear Functions_2b','dSAT_Linear Functions_3a','dSAT_Linear Functions_2a','dSAT_Linear Functions_1b',
  'dSAT_Linear Functions_1a','dSAT_Lines angles and triangles_1b','dSAT_Linear Inequalities_3a','dSAT_Linear Inequalities_2a',
  'dSAT_Lines angles and triangles_1a','dSAT_Linear Inequalities_1a','dSAT_Models and Scatterplots_1b',
  'dSAT_Lines angles and triangles_3a','dSAT_Models and Scatterplots_1a','dSAT_Lines angles and triangles_2a'
];

/**
 * A dialog that replicates the structure of the original HTML code.
 * Each test type has its own block of UI (radio buttons, dropdowns, text fields),
 * so users only see the relevant fields for that assignment.
 *
 * Note: This version auto-populates the Class ID and Student Folder ID fields by fetching
 * business details from your Go backend. The component expects a prop 'studentFirebaseID'
 * to use in that lookup.
 */
const AssignHomeworkDialog = ({ open, onClose, onSubmit, studentFirebaseID }) => {
  // Overall test type state.
  const [testType, setTestType] = useState('ACT');

  // === ACT state ===
  const [actSection, setActSection] = useState('english');
  const [actForm, setActForm] = useState('F07');
  const [actWork, setActWork] = useState('');
  const [actTimed, setActTimed] = useState(false);
  const [actNotes, setActNotes] = useState(true);
  const [actDate, setActDate] = useState('');

  // === SAT state ===
  const [satSection, setSatSection] = useState('reading');
  const [satForm, setSatForm] = useState('sat1');
  const [satWork, setSatWork] = useState('');
  const [satTimed, setSatTimed] = useState(false);
  const [satNotes, setSatNotes] = useState(true);
  const [satDate, setSatDate] = useState('');

  // === DSAT state ===
  const [dsatSection, setDsatSection] = useState('reading1');
  const [dsatForm, setDsatForm] = useState('dsat1');
  const [dsatWork, setDsatWork] = useState('');
  const [dsatTimed, setDsatTimed] = useState(false);
  const [dsatNotes, setDsatNotes] = useState(true);
  const [dsatDate, setDsatDate] = useState('');

  // === DSAT Quiz state ===
  const [dsatQuizSection, setDsatQuizSection] = useState('Reading'); // "Reading" or "Math"
  const [dsatQuizName, setDsatQuizName] = useState('');
  const [dsatQuizDate, setDsatQuizDate] = useState('');

  // Additional data from the student's business details.
  const [classID, setClassID] = useState('');
  const [studentFolderID, setStudentFolderID] = useState('');

  // Reset form fields when the dialog opens.
  useEffect(() => {
    if (open) {
      setTestType('ACT');

      // ACT
      setActSection('english');
      setActForm('F07');
      setActWork('');
      setActTimed(false);
      setActNotes(true);
      setActDate('');

      // SAT
      setSatSection('reading');
      setSatForm('sat1');
      setSatWork('');
      setSatTimed(false);
      setSatNotes(true);
      setSatDate('');

      // DSAT
      setDsatSection('reading1');
      setDsatForm('dsat1');
      setDsatWork('');
      setDsatTimed(false);
      setDsatNotes(true);
      setDsatDate('');

      // DSAT Quiz
      setDsatQuizSection('Reading');
      setDsatQuizName('');
      setDsatQuizDate('');

      // Extra fields
      setClassID('');
      setStudentFolderID('');
    }
  }, [open]);

  // When the dialog opens and we have a studentFirebaseID, fetch business details.
  useEffect(() => {
    if (open && studentFirebaseID) {
      async function fetchBusinessDetails() {
        try {
          const response = await fetch(
            `/api/tutor/get-business-details?firebase_id=${encodeURIComponent(studentFirebaseID)}`,
            {
              method: 'GET',
              headers: { 'Content-Type': 'application/json' }
            }
          );
          if (!response.ok) {
            throw new Error('Failed to fetch business details');
          }
          const result = await response.json();
          if (result.business) {
            setClassID(result.business.classroom_id || '');
            setStudentFolderID(result.business.student_folder_id || '');
          }
        } catch (error) {
          console.error('Error fetching business details:', error);
        }
      }
      fetchBusinessDetails();
    }
  }, [open, studentFirebaseID]);

  // Build the payload and call onSubmit.
  const handleSubmit = () => {
    let payload = {
      class_id: classID.trim(),
      student_folder_id: studentFolderID.trim(),
    };

    if (!payload.class_id || !payload.student_folder_id) {
      alert('Please fill Class ID and Student Folder ID.');
      return;
    }

    switch (testType) {
      case 'ACT':
        if (!actDate || !actWork) {
          alert('Fill out all required ACT fields (Problems/Passages and Due Date).');
          return;
        }
        payload = {
          ...payload,
          test: 'ACT',
          section: actSection,
          form: actForm,
          work: actWork,
          timed: actTimed,
          notes: actNotes,
          date: actDate
        };
        break;
      case 'SAT':
        if (!satDate || !satWork) {
          alert('Fill out all required SAT fields (Problems/Passages and Due Date).');
          return;
        }
        payload = {
          ...payload,
          test: 'SAT',
          section: satSection,
          form: satForm,
          work: satWork,
          timed: satTimed,
          notes: satNotes,
          date: satDate
        };
        break;
      case 'DSAT':
        if (!dsatDate || !dsatWork) {
          alert('Fill out all required DSAT fields (Problems/Passages and Due Date).');
          return;
        }
        payload = {
          ...payload,
          test: 'DSAT',
          section: dsatSection,
          form: dsatForm,
          work: dsatWork,
          timed: dsatTimed,
          notes: dsatNotes,
          date: dsatDate
        };
        break;
      case 'DSATquiz':
        if (!dsatQuizDate || !dsatQuizName) {
          alert('Please select the quiz name and due date.');
          return;
        }
        payload = {
          ...payload,
          test: 'DSATquiz',
          section: dsatQuizSection, // "Reading" or "Math"
          quiz: dsatQuizName,
          date: dsatQuizDate
        };
        break;
      default:
        alert('Invalid test type.');
        return;
    }

    onSubmit(payload);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Assign Homework</DialogTitle>
      <DialogContent>
        {/* TEST TYPE RADIO GROUP */}
        <FormControl component="fieldset" sx={{ mt: 2 }}>
          <FormLabel component="legend">Homework Type</FormLabel>
          <RadioGroup
            row
            value={testType}
            onChange={(e) => setTestType(e.target.value)}
          >
            <FormControlLabel value="ACT" control={<Radio />} label="ACT" />
            <FormControlLabel value="SAT" control={<Radio />} label="SAT" />
            <FormControlLabel value="DSAT" control={<Radio />} label="DSAT" />
            <FormControlLabel value="DSATquiz" control={<Radio />} label="DSAT Quiz" />
          </RadioGroup>
        </FormControl>

        {/* ACT BLOCK */}
        {testType === 'ACT' && (
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6">ACT Homework Info</Typography>
            <FormControl component="fieldset" sx={{ mt: 2 }}>
              <FormLabel component="legend">Section</FormLabel>
              <RadioGroup
                row
                value={actSection}
                onChange={(e) => setActSection(e.target.value)}
              >
                <FormControlLabel value="english" control={<Radio />} label="English" />
                <FormControlLabel value="math" control={<Radio />} label="Math" />
                <FormControlLabel value="reading" control={<Radio />} label="Reading" />
                <FormControlLabel value="science" control={<Radio />} label="Science" />
                <FormControlLabel
                  value="pt"
                  control={<Radio />}
                  label="Practice Test (no touchy)"
                />
              </RadioGroup>
            </FormControl>

            <FormControl fullWidth margin="normal">
              <InputLabel>Choose Test</InputLabel>
              <Select
                value={actForm}
                label="Choose Test"
                onChange={(e) => setActForm(e.target.value)}
              >
                {ACT_FORMS.map((f) => (
                  <MenuItem key={f} value={f}>
                    {f}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              fullWidth
              margin="normal"
              label="Problems/Passages"
              value={actWork}
              onChange={(e) => setActWork(e.target.value)}
            />
            <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={actTimed}
                    onChange={(e) => setActTimed(e.target.checked)}
                  />
                }
                label="Timed?"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={actNotes}
                    onChange={(e) => setActNotes(e.target.checked)}
                  />
                }
                label="Notes?"
              />
            </Box>
            <TextField
              fullWidth
              margin="normal"
              label="Due Date"
              type="date"
              value={actDate}
              onChange={(e) => setActDate(e.target.value)}
              InputLabelProps={{ shrink: true }}
            />
          </Box>
        )}

        {/* SAT BLOCK */}
        {testType === 'SAT' && (
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6">SAT Homework Info</Typography>
            <FormControl component="fieldset" sx={{ mt: 2 }}>
              <FormLabel component="legend">Section</FormLabel>
              <RadioGroup
                row
                value={satSection}
                onChange={(e) => setSatSection(e.target.value)}
              >
                <FormControlLabel value="reading" control={<Radio />} label="Reading" />
                <FormControlLabel value="writing" control={<Radio />} label="Writing" />
                <FormControlLabel value="nocalc" control={<Radio />} label="No Calc" />
                <FormControlLabel value="calc" control={<Radio />} label="Calculator" />
                <FormControlLabel value="math" control={<Radio />} label="Math" />
              </RadioGroup>
            </FormControl>

            <FormControl fullWidth margin="normal">
              <InputLabel>Choose Test</InputLabel>
              <Select
                value={satForm}
                label="Choose Test"
                onChange={(e) => setSatForm(e.target.value)}
              >
                {SAT_FORMS.map((f) => (
                  <MenuItem key={f} value={f}>
                    {f}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              fullWidth
              margin="normal"
              label="Problems/Passages"
              value={satWork}
              onChange={(e) => setSatWork(e.target.value)}
            />
            <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={satTimed}
                    onChange={(e) => setSatTimed(e.target.checked)}
                  />
                }
                label="Timed?"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={satNotes}
                    onChange={(e) => setSatNotes(e.target.checked)}
                  />
                }
                label="Notes?"
              />
            </Box>
            <TextField
              fullWidth
              margin="normal"
              label="Due Date"
              type="date"
              value={satDate}
              onChange={(e) => setSatDate(e.target.value)}
              InputLabelProps={{ shrink: true }}
            />
          </Box>
        )}

        {/* DSAT BLOCK */}
        {testType === 'DSAT' && (
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6">DSAT Homework Info</Typography>
            <FormControl component="fieldset" sx={{ mt: 2 }}>
              <FormLabel component="legend">Section</FormLabel>
              <RadioGroup
                row
                value={dsatSection}
                onChange={(e) => setDsatSection(e.target.value)}
              >
                <FormControlLabel value="reading1" control={<Radio />} label="Reading 1" />
                <FormControlLabel value="reading2" control={<Radio />} label="Reading 2" />
                <FormControlLabel value="math1" control={<Radio />} label="Math 1" />
                <FormControlLabel value="math2" control={<Radio />} label="Math 2" />
              </RadioGroup>
            </FormControl>

            <FormControl fullWidth margin="normal">
              <InputLabel>Choose Test</InputLabel>
              <Select
                value={dsatForm}
                label="Choose Test"
                onChange={(e) => setDsatForm(e.target.value)}
              >
                {DSAT_FORMS.map((f) => (
                  <MenuItem key={f} value={f}>
                    {f}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              fullWidth
              margin="normal"
              label="Problems"
              value={dsatWork}
              onChange={(e) => setDsatWork(e.target.value)}
            />
            <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={dsatTimed}
                    onChange={(e) => setDsatTimed(e.target.checked)}
                  />
                }
                label="Timed?"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={dsatNotes}
                    onChange={(e) => setDsatNotes(e.target.checked)}
                  />
                }
                label="Notes?"
              />
            </Box>
            <TextField
              fullWidth
              margin="normal"
              label="Due Date"
              type="date"
              value={dsatDate}
              onChange={(e) => setDsatDate(e.target.value)}
              InputLabelProps={{ shrink: true }}
            />
          </Box>
        )}

        {/* DSAT QUIZ BLOCK */}
        {testType === 'DSATquiz' && (
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6">DSAT Quiz Info</Typography>
            <FormControl component="fieldset" sx={{ mt: 2 }}>
              <FormLabel component="legend">Section</FormLabel>
              <RadioGroup
                row
                value={dsatQuizSection}
                onChange={(e) => {
                  setDsatQuizSection(e.target.value);
                  setDsatQuizName('');
                }}
              >
                <FormControlLabel value="Reading" control={<Radio />} label="Reading" />
                <FormControlLabel value="Math" control={<Radio />} label="Math" />
              </RadioGroup>
            </FormControl>

            {dsatQuizSection === 'Reading' && (
              <FormControl fullWidth margin="normal">
                <InputLabel>Choose Reading Quiz</InputLabel>
                <Select
                  value={dsatQuizName}
                  label="Choose Reading Quiz"
                  onChange={(e) => setDsatQuizName(e.target.value)}
                >
                  {DSAT_QUIZ_READING.map((quiz) => (
                    <MenuItem key={quiz} value={quiz}>
                      {quiz}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}
            {dsatQuizSection === 'Math' && (
              <FormControl fullWidth margin="normal">
                <InputLabel>Choose Math Quiz</InputLabel>
                <Select
                  value={dsatQuizName}
                  label="Choose Math Quiz"
                  onChange={(e) => setDsatQuizName(e.target.value)}
                >
                  {DSAT_QUIZ_MATH.map((quiz) => (
                    <MenuItem key={quiz} value={quiz}>
                      {quiz}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}

            <TextField
              fullWidth
              margin="normal"
              label="Due Date"
              type="date"
              value={dsatQuizDate}
              onChange={(e) => setDsatQuizDate(e.target.value)}
              InputLabelProps={{ shrink: true }}
            />
          </Box>
        )}

        {/* CLASS ID & STUDENT FOLDER ID (shared across all) */}
        <Box sx={{ mt: 3 }}>
          <TextField
            fullWidth
            label="Class ID"
            margin="normal"
            value={classID}
            onChange={(e) => setClassID(e.target.value)}
          />
          <TextField
            fullWidth
            label="Student Folder ID"
            margin="normal"
            value={studentFolderID}
            onChange={(e) => setStudentFolderID(e.target.value)}
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" onClick={handleSubmit}>
          Submit
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default AssignHomeworkDialog;
