#!/bin/bash
#
# Fetch price quotes for DFA funds
#
# 1. Open the fund in a browser with
#    http://idata.fundata.com/mutualfunds/Search.aspx?SearchTerm=dfa832
# 2. Click on the fund link to view summary
# 3. Copy the fund ID from the summary link
# 4. Add a new line for the fund below using the fund ID
#
quote () {
  curl -s http://idata.fundata.com/MutualFunds/FundSnapshot.aspx?IID=$1 |\
  sed -n 's/.*<span id="ctl00_MainContent_txtNavps">\$\(.*\)<\/span>.*/\1/p'
}

echo P $(date +%Y/%m/%d ) DFA50 $(quote 407115) USD
echo P $(date +%Y/%m/%d ) DFA60 $(quote 308638) USD
echo P $(date +%Y/%m/%d ) DFAEQ $(quote 308598) USD
