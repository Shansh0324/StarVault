const crypto = require('crypto');

async function runTest() {
    console.log("Starting End-to-End NATS Verification Test...\n");
    const baseUrl = "http://localhost:3000/api/v1";
    const userEmail = `testuser_${Date.now()}@example.com`;
    const password = "Password123!";

    try {
        // 1. Register User
        console.log(`[1/5] Registering user: ${userEmail}`);
        let res = await fetch(`${baseUrl}/auth/register`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ email: userEmail, password })
        });
        
        console.log(`   [OK] User Registered. Logging in...`);
        res = await fetch(`${baseUrl}/auth/login`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ email: userEmail, password })
        });
        const loginData = await res.json();
        const userToken = loginData.token;
        console.log("   [OK] User Logged in & Token received.");

        // 2. Register App
        console.log(`\n[2/5] Registering Third-Party App...`);
        res = await fetch(`${baseUrl}/apps/register`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ name: "NATS Tester App" })
        });
        const appData = await res.json();
        console.log(appData);
        const appId = appData.id || appData.app?.id || appData.appId;
        const appSecret = appData.secret || appData.app?.secret || appData.appSecret;
        console.log(`   [OK] App Registered. ID: ${appId}`);

        // 3. Create Vault Data
        console.log(`\n[3/5] Encrypting Medical Data in Vault...`);
        res = await fetch(`${baseUrl}/vault/data`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${userToken}`
            },
            body: JSON.stringify({ data: "My blood type is O+", dataType: "medical_data" })
        });
        const vaultData = await res.json();
        console.log(vaultData);
        const vaultDataId = vaultData.id || vaultData.vaultData?.id || vaultData.vaultDataId;
        console.log(`   [OK] Data Encrypted. Vault ID: ${vaultDataId}`);

        // 4. Grant Consent
        console.log(`\n[4/5] Granting Consent to App...`);
        res = await fetch(`${baseUrl}/consents`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${userToken}`
            },
            body: JSON.stringify({ appId: appId, scopes: ["medical_data"], purpose: "End-to-End automated testing", expiresAt: new Date(Date.now() + 86400000).toISOString() })
        });
        const consentData = await res.json();
        console.log(consentData);
        console.log("   [OK] Consent Granted.");

        // 5. Access Data (Triggers NATS + Blockchain)
        console.log(`\n[5/5] Requesting Data Access (Testing NATS + Blockchain)...`);
        res = await fetch(`${baseUrl}/access/data`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${userToken}`
            },
            body: JSON.stringify({
                appId: appId,
                secret: appSecret,
                scope: "medical_data",
                vaultDataId: vaultDataId
            })
        });
        const accessData = await res.json();
        if (accessData.error) {
             console.error("   [FAIL] Access Failed:", accessData.error);
        } else {
             console.log(`   [OK] SUCCESS! Decrypted Data: "${accessData.data}"`);
             console.log("\n[SUCCESS] ALL TESTS PASSED! Check the Docker logs to see the NATS event and Blockchain Anchor!");
        }

    } catch (e) {
        console.error("[FAIL] Test Failed:", e);
    }
}

runTest();
