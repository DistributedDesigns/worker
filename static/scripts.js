function SendPost(theCommand) {
  var url = '/push';
  var request = new XMLHttpRequest();
  request.open("POST", url, true);

  var jsonPayload = JSON.stringify({command: theCommand})
  request.send(jsonPayload);
}

function appendLog(item) {
    var log = document.getElementById("log");
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
        log.scrollTop = log.scrollHeight - log.clientHeight;
    }
}

function doAuth(rawData) {
  localStorage["user"] = localStorage["user"] || document.getElementById("userid").value
  document.getElementById('loginview').style.display = 'none';
  document.getElementById('authview').style.display = 'block';
  var conn;
  var msg = document.getElementById("msg");
  if (window["WebSocket"]) {
      conn = new WebSocket("ws://" + document.location.host + "/ws");
      conn.onopen = function (evt) {
        console.log("CON OPEN")
        conn.send(localStorage["user"])
      }
      conn.onclose = function (evt) {
          var item = document.createElement("div");
          item.innerHTML = "<b>Connection closed.</b>";
          appendLog(item);
      };
      conn.onmessage = function (evt) {
          jsonData = JSON.parse(evt.data)

          message = jsonData["message"]
          if (message !== "") {
            var item = document.createElement("div");
            item.innerText = message;
            appendLog(item)
          }

          userData = jsonData["account"]
          document.getElementById("balance").innerHTML = userData["balance"]

          buyContainer = document.createElement("div")
          userData["pendingBuys"].map(function(elem) {
            var item = document.createElement("div");
            item.innerText = "{" + elem.amount + "," + elem.expiresAt + "," + elem.stock + "}"
            buyContainer.appendChild(item)
          })
          document.getElementById("pendingbuys").innerHTML = buyContainer.innerHTML

          sellContainer = document.createElement("div")
          userData["pendingSells"].map(function(elem) {
            var item = document.createElement("div");
            item.innerText = "{" + elem.amount + "," + elem.expiresAt + "," + elem.stock + "}"
            sellContainer.appendChild(item)
          })
          document.getElementById("pendingsells").innerHTML = sellContainer.innerHTML
          
          portfolioContainer = document.createElement("div")
          portfolio = userData["portfolio"]
          for (var key in portfolio) {
            if (portfolio.hasOwnProperty(key)) {
              var item = document.createElement("div");
              item.innerText = key + " -> " + portfolio[key];
              portfolioContainer.appendChild(item)
            }
          }
          document.getElementById("portfolio").innerHTML = portfolioContainer.innerHTML

          autoTxContainer = document.createElement("div")
          userData["pendingATX"].map(function(elem) {
            console.log(elem)
          })
      };
  } else {
      var item = document.createElement("div");
      item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
      appendLog(item);
  }
  console.log("Auth Successful")
}
function Auth() {
  var request = new XMLHttpRequest();
  request.open('POST', '/auth', true);

  request.onload = function() {
    if (request.status >= 200 && request.status < 400) {
      // Success!
      // var data = request.responseText;
      // console.log(data)
      doAuth(request.responseText)
    } else {
      // We reached our target server, but it returned an error
      console.log("Failed to Auth user")
    }
  };

  request.onerror = function() {
    // There was a connection error of some sort
  };

  request.send(JSON.stringify({user: document.getElementById("userid").value, pass: document.getElementById("pwd").value}));
}
function Create() {
  var request = new XMLHttpRequest();
  request.open('POST', '/create', true);

  request.onload = function() {
    if (request.status >= 200 && request.status < 400) {
      // Success!
      // var data = request.responseText;
      // console.log(data)
      doAuth(request.responseText)
      console.log("Create Successful")
    } else {
      // We reached our target server, but it returned an error
      console.log("Failed to create user")
    }
  };

  request.onerror = function() {
    // There was a connection error of some sort
  };

  request.send(JSON.stringify({user: document.getElementById("userid").value, pass: document.getElementById("pwd").value}));
}

function AmountCheck(theAmount){
  var reg = /^\d+\.?(\d\d)?$/
  if (reg.test(document.getElementById("addamount").value)) {
    return true
  }
  else{
    var item = document.createElement("div");
    item.innerText = "Invalid input for amount. Expected format example: '100.00'"
    appendLog(item)
    return false
  }
}

function SymbolCheck(theStockSymbol){
  var reg = /^[A-Z][A-Z][A-Z]$/
  if (reg.test(theStockSymbol)) {
    return true
  }
  else{
    var item = document.createElement("div");
    item.innerText = "Invalid input for stock symbol. Expected 3 letter upper string.  Format example: 'ASS'"
    appendLog(item)
    return false
  }
}

function Logout() {
  localStorage.clear()
  location.reload()
}
function Add() {
  if (AmountCheck(document.getElementById("addamount").value)) {
    SendPost("ADD,"+localStorage["user"]+","+document.getElementById("addamount").value)
    console.log("Add")
  }
}
function Quote() {
  if (SymbolCheck(document.getElementById("quotestocksymbol").value)) {
    SendPost("QUOTE,"+localStorage["user"]+","+document.getElementById("quotestocksymbol").value.toUpperCase())
    console.log("Quote")
  }  
}
function Buy() {
  if(AmountCheck(document.getElementById("buyamount").value) && SymbolCheck(document.getElementById("buystocksymbol").value)){
    SendPost("BUY,"+localStorage["user"]+","+document.getElementById("buystocksymbol").value.toUpperCase()+","+document.getElementById("buyamount").value)
    console.log("Buy")
  }
}
function CommitBuy() {
  SendPost("COMMIT_BUY,"+localStorage["user"])
  console.log("Commit Buy")
}
function CancelBuy() {
  SendPost("CANCEL_BUY,"+localStorage["user"])
  console.log("Cancel Buy")
}
function Sell() {
  if(AmountCheck(document.getElementById("sellamount").value) && SymbolCheck(document.getElementById("sellstocksymbol").value)){    
    SendPost("SELL,"+localStorage["user"]+","+document.getElementById("sellstocksymbol").value.toUpperCase()+","+document.getElementById("sellamount").value)
    console.log("Sell")
  }
}
function CommitSell() {
  SendPost("COMMIT_SELL,"+localStorage["user"])
  console.log("Commit Sell")
}
function CancelSell() {
  SendPost("CANCEL_SELL,"+localStorage["user"])
  console.log("Cancel Sell")
}
function SetBuyAmount() {
  if(AmountCheck(document.getElementById("setbuyamount").value) && SymbolCheck(document.getElementById("setbuystocksymbol").value)){    
    SendPost("SET_BUY_AMOUNT,"+localStorage["user"]+","+document.getElementById("setbuystocksymbol").value.toUpperCase()+","+document.getElementById("setbuyamount").value)
    console.log("Set buy amount")
  }
}
function CancelSetBuy() {
  if(AmountCheck(SymbolCheck(document.getElementById("cancelbuystocksymbol").value))) {    
    SendPost("CANCEL_SET_BUY,"+localStorage["user"]+","+document.getElementById("cancelbuystocksymbol").value.toUpperCase())
    console.log("Cancel set buy")
  }
}
function SetBuyTrigger() {
  if(AmountCheck(SymbolCheck(document.getElementById("setbuytrigstocksymbol").value))) {    
    SendPost("SET_BUY_TRIGGER,"+localStorage["user"]+","+document.getElementById("setbuytrigstocksymbol").value.toUpperCase()+","+document.getElementById("setbuytrigamount").value)
    console.log("Set buy trigger")
  }
}
function SetSellAmount() {
  if(AmountCheck(document.getElementById("setsellamount").value) && SymbolCheck(document.getElementById("setsellstocksymbol").value)){    
    SendPost("SET_SELL_AMOUNT,"+localStorage["user"]+","+document.getElementById("setsellstocksymbol").value.toUpperCase()+","+document.getElementById("setsellamount").value)
    console.log("Set Sell Amount")
  }
}
function CancelSetSell() {
  if(AmountCheck(SymbolCheck(document.getElementById("cancelsellstocksymbol").value))) {    
    SendPost("CANCEL_SET_SELL,"+localStorage["user"]+","+document.getElementById("cancelsellstocksymbol").value.toUpperCase())
    console.log("Cancel set sell")
  }
}
function SetSellTrigger() {
  if(AmountCheck(SymbolCheck(document.getElementById("setselltrigstocksymbol").value))) {    
    SendPost("SET_SELL_TRIGGER,"+localStorage["user"]+","+document.getElementById("setselltrigstocksymbol").value.toUpperCase()+","+document.getElementById("setselltrigamount").value)
    console.log("Set sell trigger")
  }
}
function Dumplog() {
  var userID = localStorage["user"]
  if(userID === "admin"){
    SendPost("DUMPLOG,out.dump")
    console.log("Admin Dump")
  }
  else
  {
    SendPost("DUMPLOG,"+localStorage["user"]+","+document.getElementById("dumplogfile").value)
    console.log("User Dump")
  }
}
function DisplaySummary() {
  SendPost("DISPLAY_SUMMARY,"+localStorage["user"])
  console.log("Display summary")
}
