<?php
$SKIP_AUTH = true;
require_once "_incl/utils.php";

$error_message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['login_user_pass'])) {
        // Handle username/password login
        $username = $_POST['username'];
        $password = $_POST['password'];
        if (login($username, $password)) {
            header("Location: /index.php");
            exit();
        } else {
            $error_message = "Usuario o contraseña incorrectos.";
        }
    } elseif (isset($_POST['qr_data'])) {
        // Handle QR code login
        if (login_with_qr($_POST['qr_data'])) {
            header("Location: /index.php");
            exit();
        } else {
            $error_message = "Código QR inválido o caducado.";
        }
    }
}

require_once "_incl/pre-body.php";
?>

<div id="login-container">
    <h1>Iniciar Sesión</h1>

    <?php if ($error_message): ?>
        <p style="color: red;"><?php echo $error_message; ?></p>
    <?php endif; ?>

    <div id="login-method-selector">
        <button id="btn-user-pass" class="active"><img loading="lazy" class="picto" src="/static/pictos/contraseña-recordar.png"><br> Usuario/Contraseña</button>
        <button id="btn-qr"><img loading="lazy" class="picto" src="/static/pictos/QR.png"><br> Codigo QR</button>
    </div>

    <form id="login-form-user-pass" method="POST" action="login.php">
        <div class="form-group">
            <label for="username"><img loading="lazy" class="picto" src="/static/pictos/nombre.png"></label>
            <input type="text" id="username" name="username" autofocus required placeholder="Usuario">
        </div>
        <div class="form-group">
            <label for="password"><img loading="lazy" class="picto" src="/static/pictos/contraseña.png"></label>
            <input type="password" id="password" name="password" required placeholder="Contraseña">
        </div>
        <button type="submit" name="login_user_pass"><img loading="lazy" class="picto" src="/static/pictos/abrir-cerradura.png"><br> Entrar</button>
    </form>

    <div id="login-form-qr" style="display: none;">
        <p>Escanea el código QR para iniciar sesión.</p>
        <div id="qr-scanner-container">
        </div>
        <div id="qr-result" style="display:none;"></div>
    </div>
</div>

<script src="https://unpkg.com/html5-qrcode" type="text/javascript"></script>
<script>
document.addEventListener("DOMContentLoaded", () => {
    const btnUserPass = document.getElementById('btn-user-pass');
    const btnQr = document.getElementById('btn-qr');
    const formUserPass = document.getElementById('login-form-user-pass');
    const formQr = document.getElementById('login-form-qr');
    const qrResult = document.getElementById('qr-result');
    
    let html5QrCode;

    function onScanSuccess(decodedText, decodedResult) {
        // handle the scanned code as you like, for example:
        console.log(`Code matched = ${decodedText}`, decodedResult);
        
        if (html5QrCode && html5QrCode.isScanning) {
            html5QrCode.stop().then((ignore) => {
                // QR Code scanning is stopped.
            }).catch((err) => {
                // Stop failed, handle it.
            });
        }

        qrResult.style.display = 'block';
        qrResult.textContent = 'QR detectado. Procesando...';
        
        // Create a form and submit the data
        const qrForm = document.createElement('form');
        qrForm.method = 'POST';
        qrForm.action = 'login.php';
        
        const qrDataInput = document.createElement('input');
        qrDataInput.type = 'hidden';
        qrDataInput.name = 'qr_data';
        qrDataInput.value = decodedText;
        qrForm.appendChild(qrDataInput);
        
        document.body.appendChild(qrForm);
        qrForm.submit();
    }

    function onScanFailure(error) {
        // handle scan failure, usually better to ignore and keep scanning.
        // for example:
        // console.warn(`Code scan error = ${error}`);
    }

    function startQrScanner() {
        if (!html5QrCode) {
            html5QrCode = new Html5Qrcode("qr-scanner-container");
        }
        
        const config = { fps: 10, qrbox: { width: 250, height: 250 } };
        
        // If you want to prefer front camera
        html5QrCode.start({ facingMode: "environment" }, config, onScanSuccess, onScanFailure)
        .catch(err => {
            console.error("Error starting QR scanner", err);
            alert("No se pudo iniciar el escáner QR. Asegúrate de dar permisos para la cámara.");
        });
    }

    function stopQrScanner() {
        if (html5QrCode && html5QrCode.isScanning) {
            html5QrCode.stop().catch(err => console.log("Failed to stop scanner cleanly", err));
        }
    }

    btnUserPass.addEventListener('click', () => {
        stopQrScanner();
        formUserPass.style.display = 'block';
        formQr.style.display = 'none';
        btnUserPass.classList.add('active');
        btnQr.classList.remove('active');
    });

    btnQr.addEventListener('click', () => {
        formUserPass.style.display = 'none';
        formQr.style.display = 'block';
        btnQr.classList.add('active');
        btnUserPass.classList.remove('active');
        startQrScanner();
    });
});
</script>

<?php require_once "_incl/post-body.php"; ?>
