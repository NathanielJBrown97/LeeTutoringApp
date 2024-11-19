function loadHomeworkIntoSheet(comp) {
  Logger.log("Logging homework completion");

  let dateStr = formatDate(new Date());

  // Save the original value of comp for backend
  let percentage = comp;

  // Append "%" for the spreadsheet if comp is not "N/A"
  if (comp != "N/A") {
    comp += "%";
  }

  SpreadsheetApp.getActiveSpreadsheet().getSheetByName("homework completion").appendRow([dateStr, comp]);

  // Get student name from the "Profile" sheet
  let studentName = SpreadsheetApp.getActiveSpreadsheet().getSheetByName("Profile").getRange("B4").getValue();

  // Send data to the backend
  sendHomeworkCompletionToBackend(studentName, dateStr, percentage);
}




function sendHomeworkCompletionToBackend(studentName, dateStr, percentage) {
  try {
    let url = "https://agora-backend-1057197198698.us-east1.run.app/cmd/firestoreupdater/homeworkCompletion";

    let data = {
      "studentName": studentName,
      "date": dateStr,
      "percentage": percentage
    };

    let options = {
      'method': 'post',
      'contentType': 'application/json',
      'payload': JSON.stringify(data),
      'muteHttpExceptions': true
    };

    let response = UrlFetchApp.fetch(url, options);
    Logger.log("Homework completion data sent to backend");
    Logger.log(response.getContentText());
    
  } catch (error) {
    Logger.log("Failed to send homework completion data to backend");
    Logger.log(error);
  }
}