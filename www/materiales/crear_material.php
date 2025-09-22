<?php
require_once "../_incl/utils.php";
require_permission("materiales.create");
require_once "../_incl/pre-body.php";

$centro = get_selected_centro();
$materiales_path = "$RUTA_DATOS/Centros/$centro/Materiales";

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $nombre = $_POST['nombre'] ?? '';
    $unidad = $_POST['unidad'] ?? 'unidad';
    $cantidad_disponible = $_POST['cantidad_disponible'] ?? 0;
    $cantidad_minima = $_POST['cantidad_minima'] ?? 0;
    $notas = $_POST['notas'] ?? '';
    $foto_path = '/static/pictos/material_escolar.png'; // Default photo

    $uploads_dir = "$RUTA_DATOS/Centros/$centro/Materiales/Fotos";
    if (!file_exists($uploads_dir)) {
        mkdir($uploads_dir, 0777, true);
    }

    if (isset($_FILES['foto']) && $_FILES['foto']['error'] == UPLOAD_ERR_OK) {
        $tmp_name = $_FILES['foto']['tmp_name'];
        $img_name = uniqid('img_') . '_' . basename($_FILES['foto']['name']);
        $target_file = $uploads_dir . '/' . $img_name;
        if (move_uploaded_file($tmp_name, $target_file)) {
            $foto_path = "download_image.php?centro=$centro&image=$img_name";
        }
    }

    if (!empty($nombre)) {
        $filename = strtolower(preg_replace('/[^a-zA-Z0-9_-]/', '_', $nombre)) . '.json';
        $filepath = "$materiales_path/$filename";

        $data = [
            'nombre' => $nombre,
            'foto' => $foto_path,
            'unidad' => $unidad,
            'cantidad_disponible' => $cantidad_disponible,
            'cantidad_minima' => $cantidad_minima,
            'notas' => $notas,
            'createdAt' => date('c')
        ];

        if (file_put_contents($filepath, json_encode($data, JSON_PRETTY_PRINT))) {
            header("Location: index.php");
            exit();
        } else {
            $error = "No se pudo guardar el material.";
        }
    } else {
        $error = "El nombre es obligatorio.";
    }
}
if (!file_exists($materiales_path)) {
    if (!mkdir($materiales_path, 0777, true)) {
        $error = "Error: No se pudo crear el directorio de materiales.";
    }
}

?>

<h1>Añadir Material</h1>

<?php if (isset($error)): ?>
    <p style="color: red;"><?php echo $error; ?></p>
<?php endif; ?>

<form method="POST" enctype="multipart/form-data">
    <fieldset>
        <legend>Nuevo Material</legend>

        <label for="nombre">Nombre del material:</label><br>
        <input type="text" id="nombre" name="nombre" required><br><br>

        <label for="foto">Foto:<br>
        <img id="fotoPreview" src="/static/pictos/material_escolar.png" style="max-height: 200px; max-width: 100%; border: 1px solid #ccc; padding: 10px; border-radius: 5px; cursor: pointer;"><br>
        </label>

        <input style="display: none;" type="file" id="foto" name="foto" accept="image/*"><br><br>
        
        <script>
            const fotoInput = document.getElementById('foto');
            const fotoPreview = document.getElementById('fotoPreview');
            const fotoLabel = document.querySelector('label[for="foto"]');

            fotoLabel.addEventListener('click', () => fotoInput.click());

            fotoInput.addEventListener('change', function() {
                const file = this.files[0];
                if (file) {
                    const reader = new FileReader();
                    reader.onload = function(e) {
                        fotoPreview.src = e.target.result;
                    }
                    reader.readAsDataURL(file);
                } else {
                    fotoPreview.src = '/static/pictos/material_escolar.png';
                }
            });
        </script>

        <label for="unidad">Unidad de medida:</label><br>
        <select id="unidad" name="unidad">
            <option value="unidad">Unidad</option>
            <option value="caja">Caja</option>
            <option value="paquete">Paquete</option>
            <option value="docena">Docena</option>
            <optgroup label="Metrico">
                <option value="gramos">Gramos (g)</option>
                <option value="kilogramos">Kilogramos (kg)</option>
                <option value="litros">Litros (l)</option>
                <option value="mililitros">Mililitros (ml)</option>
                <option value="milimetros">Milímetros (mm)</option>
                <option value="centimetros">Centímetros (cm)</option>
                <option value="metros">Metros (m)</option>
            </optgroup>
        </select><br><br>

        <label for="cantidad_disponible">Cantidad disponible:</label><br>
        <input type="number" id="cantidad_disponible" name="cantidad_disponible" min="0" step="1" value="0"><br><br>

        <label for="cantidad_minima">Cantidad mínima:</label><br>
        <input type="number" id="cantidad_minima" name="cantidad_minima" min="0" step="1" value="0"><br><br>

        <label for="notas">Notas adicionales:</label><br>
        <textarea id="notas" name="notas" rows="4" cols="50"></textarea><br><br>

        <br>
        <button type="submit">Crear Material</button>
    </fieldset>
</form>

<style>
    .picto-selector {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
    }

    .picto-selector label {
        cursor: pointer;
        border: 2px solid transparent;
        border-radius: 5px;
        padding: 5px;
    }

    .picto-selector input[type="radio"] {
        display: none;
    }

    .picto-selector input[type="radio"]:checked+img {
        border: 2px solid blue;
        border-radius: 5px;
    }
</style>

<?php
require_once "../_incl/post-body.php";
?>