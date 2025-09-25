<?php
require_once "../_incl/pre-body.php";

// Get activity ID and global flag from URL parameters
$id = $_GET['id'] ?? '';
$is_global = ($_GET['global'] ?? '0') === '1';

if (empty($id)) {
    header("Location: index.php");
    exit();
}

$centro = $_SESSION['centro'];
$aula = $_SESSION['aula'];

// Determine the activity directory
if ($is_global) {
    $activity_dir = "$RUTA_DATOS/Actividades/_Global";
} else {
    $activity_dir = "$RUTA_DATOS/Actividades/$centro/$aula";
}

// Load the activity
$activity_file = "$activity_dir/$id.json";
if (!file_exists($activity_file)) {
    echo "<p style='color: red;'>Actividad no encontrada.</p>";
    echo "<a href='index.php' class='button'>Volver a la lista</a>";
    require_once "../_incl/post-body.php";
    exit();
}

$activity = json_decode(file_get_contents($activity_file), true);
if (!$activity) {
    echo "<p style='color: red;'>Error al cargar la actividad.</p>";
    echo "<a href='index.php' class='button'>Volver a la lista</a>";
    require_once "../_incl/post-body.php";
    exit();
}

// Add global flag to activity data
$activity['_global'] = $is_global;
?>

<div style="margin-bottom: 20px;">
    <a href="index.php" class="button">
        <img loading="lazy" class="picto" src="/static/pictos/salir.png"><br>Volver a la lista
    </a>
</div>

<div style="background-color: lightcyan; padding: 20px; margin-top: 10px; border-radius: 25px; border: 2px solid black;">
    <h1><?php echo htmlspecialchars($activity['title']); ?></h1>
    
    <?php if ($is_global && isset($activity["is_shared_from"]) && $activity["is_shared_from"] != ""): ?>
        <p style="margin-bottom: 15px;">
            <img loading="lazy" class="picto" src="/static/pictos/compartir2.png">
            <span style="font-weight: bold; vertical-align: top; display: inline-block;">
                Compartido por: <br>
                <span style="font-size: large; color: #2c5aa0;"><?php echo htmlspecialchars(iso_to_es($activity['is_shared_from'])); ?></span>
            </span>
        </p>
    <?php endif; ?>

    <div style="display: flex; flex-wrap: wrap; gap: 20px; margin-bottom: 20px;">
        <div>
            <p>
                <img loading="lazy" class="picto" src="/static/pictos/dia.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;">
                    Fecha: <br>
                    <span style="font-size: x-large;"><?php echo htmlspecialchars(iso_to_es($activity['start'])); ?></span>
                </span>
            </p>
        </div>
        
        <div>
            <p>
                <img loading="lazy" class="picto" src="/static/pictos/inicio.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;">
                    Inicio: <br>
                    <span style="font-size: x-large;"><?php echo htmlspecialchars(explode("T", $activity['start'])[1]); ?></span>
                </span>
            </p>
        </div>
        
        <div>
            <p>
                <img loading="lazy" class="picto" src="/static/pictos/fin.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;">
                    Fin: <br>
                    <span style="font-size: x-large;"><?php echo htmlspecialchars(explode("T", $activity['end'])[1]); ?></span>
                </span>
            </p>
        </div>
    </div>

    <?php if (!empty($activity['description'])): ?>
        <div style="margin-bottom: 20px;">
            <h3>
                <img loading="lazy" class="picto" src="/static/pictos/descripcion.png">
                Descripción:
            </h3>
            <div style="background-color: white; padding: 15px; border-radius: 10px; border: 1px solid #ccc;">
                <?php echo nl2br(htmlspecialchars($activity['description'])); ?>
            </div>
        </div>
    <?php endif; ?>

    <?php if (!empty($activity['url'])): ?>
        <div style="margin-bottom: 15px;">
            <h3>
                <img loading="lazy" class="picto" src="/static/pictos/enlace.png">
                Enlace:
            </h3>
            <a href="<?php echo htmlspecialchars($activity['url']); ?>" target="_blank" 
               style="word-break: break-all; color: #2c5aa0; text-decoration: underline;">
                <?php echo htmlspecialchars($activity['url']); ?>
            </a>
        </div>
    <?php endif; ?>

    <?php if (!empty($activity['meet'])): ?>
        <div style="margin-bottom: 15px;">
            <h3>
                <img loading="lazy" class="picto" src="/static/pictos/videollamada.png">
                Videollamada:
            </h3>
            <div style="background-color: white; padding: 10px; border-radius: 10px; border: 1px solid #ccc;">
                <?php echo htmlspecialchars($activity['meet']); ?>
            </div>
        </div>
    <?php endif; ?>

    <div style="margin-top: 30px; text-align: center;">
        <a href="editar_actividad.php?id=<?php echo urlencode($activity['id']); ?>&global=<?php echo $is_global ? '1' : '0'; ?>"
            class="button">
            <img loading="lazy" class="picto" src="/static/pictos/escribir.png"><br>Editar
        </a>
        
        <a href="eliminar_actividad.php?id=<?php echo urlencode($activity['id']); ?>&global=<?php echo $is_global ? '1' : '0'; ?>"
            onclick="return confirm('¿Estás seguro de que deseas eliminar esta actividad?');" 
            class="button rojo">
            <img loading="lazy" class="picto" src="/static/pictos/borrar.png"><br>Eliminar
        </a>
    </div>
</div>

<?php
require_once "../_incl/post-body.php";
?>
