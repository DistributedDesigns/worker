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
function Logout() {
  localStorage.clear()
  location.reload()
}
function Add() {
  SendPost("ADD,"+localStorage["user"]+","+document.getElementById("addamount").value)
  console.log("Add")
}
function Quote() {
  SendPost("QUOTE,"+localStorage["user"]+","+document.getElementById("quotestocksymbol").value.toUpperCase())
  console.log("Quote")
}
function Buy() {
  SendPost("BUY,"+localStorage["user"]+","+document.getElementById("buystocksymbol").value.toUpperCase()+","+document.getElementById("buyamount").value)
  console.log("Buy")
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
  SendPost("SELL,"+localStorage["user"]+","+document.getElementById("sellstocksymbol").value.toUpperCase()+","+document.getElementById("sellamount").value)
  console.log("Sell")
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
  SendPost("SET_BUY_AMOUNT,"+localStorage["user"]+","+document.getElementById("setbuystocksymbol").value.toUpperCase()+","+document.getElementById("setbuyamount").value)
  console.log("Set buy amount")
}
function CancelSetBuy() {
  SendPost("CANCEL_SET_BUY,"+localStorage["user"]+","+document.getElementById("cancelbuystocksymbol").value.toUpperCase())
  console.log("Cancel set buy")
}
function SetBuyTrigger() {
  SendPost("SET_BUY_TRIGGER,"+localStorage["user"]+","+document.getElementById("setbuytrigstocksymbol").value.toUpperCase()+","+document.getElementById("setbuytrigamount").value)
  console.log("Set buy trigger")
}
function SetSellAmount() {
  SendPost("SET_SELL_AMOUNT,"+localStorage["user"]+","+document.getElementById("setsellstocksymbol").value.toUpperCase()+","+document.getElementById("setsellamount").value)
  console.log("Set Sell Amount")
}
function CancelSetSell() {
  SendPost("CANCEL_SET_SELL,"+localStorage["user"]+","+document.getElementById("cancelsellstocksymbol").value.toUpperCase())
  console.log("Cancel set sell")
}
function SetSellTrigger() {
  SendPost("SET_SELL_TRIGGER,"+localStorage["user"]+","+document.getElementById("setselltrigstocksymbol").value.toUpperCase()+","+document.getElementById("setselltrigamount").value)
  console.log("Set sell trigger")
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
