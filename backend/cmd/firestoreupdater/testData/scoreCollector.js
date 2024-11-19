// in logScore(data):
ss.getSheetByName("Test Data").appendRow(row.slice(2));
// after above line
// Add API call to Firebase to send data
sendTestDataToBackend(data);


function sendTestDataToBackend(data) {
    try {
      let ss = SpreadsheetApp.getActiveSpreadsheet();
      let name = ss.getSheetByName("data").getRange("A2").getValue();
  
      let payload = {
        "studentName": name,
        "date": data["date"],
        "test": data["test"],
        "quality": data["quality"],
        "baseline": data["baseline"],
        "scores": data["scores"]
      };
  
      let url = "https://agora-backend-1057197198698.us-east1.run.app/cmd/firestoreupdater/testData";
      let options = {
        'method': 'post',
        'contentType': 'application/json',
        'payload': JSON.stringify(payload),
        'muteHttpExceptions': true
      };
  
      let response = UrlFetchApp.fetch(url, options);
      Logger.log("Test data sent to backend");
      Logger.log(response.getContentText());
    } catch (error) {
      Logger.log("Failed to send test data to backend");
      Logger.log(error);
    }
  }