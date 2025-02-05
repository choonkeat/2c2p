<?php
    //Merchant Account Information 
	$merchantID = "JT";				//Get MerchantID when opening account with 2C2P
	$secretKey = "QEapqL2CigJE";		//Get SecretKey from 2C2P PGW Dashboard


    //Transaction Information
    $uniqueTransactionCode = "Invoice".time();
	$desc = "1 room for 2 nights";
	$amt = "000000010010";			//12 digit format
	$currencyCode = "764";			//Ref: http://en.wikipedia.org/wiki/ISO_4217
	
    //Cardholder Information
    $cardholderName = "John Doe";
    $country = "SG";
    $request3DS="Y";
    $storeCard= "Y";
    
    $paymentChannel = "";
    
        //Request Information
    $timeStamp = time();
    $apiVersion = "9.4";
  
    $encryptedCardInfo = $_POST['encryptedCardInfo'];                            //Retrieve encrypted credit card data
 
    //Construct signature string
    $stringToHash = $apiVersion . $timeStamp . $merchantID . $uniqueTransactionCode . $desc . $amt . $currencyCode . $paymentChannel . $storeCardUniqueID . $panBank .
        $country . $cardholderName . $cardholderEmail . $payCategoryID . $userDefined1 . $userDefined2 . $userDefined3 . $userDefined4 . $userDefined5 . $storeCard .
        $ippTransaction . $installmentPeriod . $interestType . $recurring . $invoicePrefix . $recurringAmount . $allowAccumulate . $maxAccumulateAmt . $recurringInterval . 
        $recurringCount . $chargeNextDate . $promotion . $request3DS . $statementDescriptor . $agentCode . $channelCode . $paymentExpiry . $mobileNo . $tokenizeWithoutAuthorization . $encryptedCardInfo;
    $hash = strtoupper(hash_hmac('sha1', $stringToHash ,$secretKey, false));      //Calculate Hash Value
    
    //available in js v1.6.7:
    $maskedCardNo = $_POST['maskedCardInfo'];                                    //Masked card number (first 6 and last 4 digit of credit card number)
    $expMonth = $_POST['expMonthCardInfo'];                                      //Card expiry month
    $expYear = $_POST['expYearCardInfo'];                                        //Card expiry Year
    
    
    //Construct payment request message
    $xml = "<PaymentRequest>
    <version>$apiVersion</version>
    <timeStamp>$timeStamp</timeStamp>
    <merchantID>$merchantID</merchantID>
    <uniqueTransactionCode>$uniqueTransactionCode</uniqueTransactionCode>
    <desc>$desc</desc>
    <amt>$amt</amt>
    <currencyCode>$currencyCode</currencyCode>  
    <paymentChannel>$paymentChannel</paymentChannel>  
    <panCountry>$country</panCountry> 
    <cardholderName>$cardholderName</cardholderName>   
    <request3DS>$request3DS</request3DS>   
    <secureHash>$hash</secureHash>
    <storeCard>$storeCard</storeCard>
    <encCardData>$encryptedCardInfo</encCardData>
    </PaymentRequest>";

    $payload = base64_encode($xml);    //Convert payload to base64
    ?>
    <!-- POST method to submit the form -->
    <form action='https://demo2.2c2p.com/2C2PFrontEnd/SecurePayment/PaymentAuth.aspx' method='POST' name='paymentRequestForm'>
        Processing payment request, Do not close the browser, press back or refresh the page.
        <?php echo "<input type='hidden' name='paymentRequest' value='".$payload."'>"; ?>
    </form>
    <script language="JavaScript">
        document.paymentRequestForm.submit();
    </script>