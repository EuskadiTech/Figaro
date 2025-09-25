<?php
require_once "../_incl/pre-body.php";

// Get activities
$centro = $_SESSION['centro'];
$aula = $_SESSION['aula'];

function get_activities_from_dir($dir, $is_global = false)
{
    $activities = [];
    if (!is_dir($dir)) {
        return $activities;
    }
    $files = array_diff(scandir($dir), ['.', '..']);
    foreach ($files as $file) {
        if (pathinfo($file, PATHINFO_EXTENSION) === 'json') {
            $content = file_get_contents("$dir/$file");
            $activity = json_decode($content, true);
            $activity["_global"] = $is_global;
            if ($activity) {
                $activities[] = $activity;
            }
        }
    }
    return $activities;
}

$local_activities = get_activities_from_dir("$RUTA_DATOS/Actividades/$centro/$aula");
$global_activities = get_activities_from_dir("$RUTA_DATOS/Actividades/_Global", true);
$all_activities = array_merge($local_activities, $global_activities);

// Process search query
$search_query = trim($_GET['q'] ?? '');
$show_past = ($_GET['past'] ?? '') === 'y';

// Filter activities based on search query
if (!empty($search_query)) {
    $search_query_lower = strtolower($search_query);
    $all_activities = array_filter($all_activities, function ($activity) use ($search_query_lower) {
        $title_match = strpos(strtolower($activity['title'] ?? ''), $search_query_lower) !== false;
        $description_match = strpos(strtolower($activity['description'] ?? ''), $search_query_lower) !== false;
        return $title_match || $description_match;
    });
}

// Filter for upcoming activities (default view)
$today = new DateTime();
$today->setTime(0, 0, 0); // Set time to beginning of the day for comparison

$upcoming_activities = array_filter($all_activities, function ($activity) use ($today) {
    $start_date = new DateTime($activity['start']);
    return $start_date >= $today;
});

// Get all activities for when checkbox is enabled
$past_activities = array_filter($all_activities, function ($activity) use ($today) {
    $start_date = new DateTime($activity['start']);
    return $start_date < $today;
});

// Sort activities by start date
usort($upcoming_activities, function ($a, $b) {
    return strtotime($a['start']) - strtotime($b['start']);
});

usort($past_activities, function ($a, $b) {
    return strtotime($b['start']) - strtotime($a['start']); // Recent past activities first
});

// Determine which activities to display
if ($show_past) {
    $activities_to_display = array_merge($upcoming_activities, $past_activities);
} else {
    $activities_to_display = $upcoming_activities;
}
?>

<!-- Search Bar and Controls -->
<form method="get" style="background-color: #f9f9f9; padding: 15px; margin-bottom: 20px; border-radius: 10px; border: 1px solid #ddd;">
    <div style="margin-bottom: 10px;">
        <label for="activity-search" style="font-weight: bold; display: block; margin-bottom: 5px;">
            🔍 Buscar actividades:
        </label>
        <input type="text" name="q" id="activity-search" placeholder="Buscar por título o descripción..." 
               value="<?php echo htmlspecialchars($search_query); ?>"
               style="width: 100%; padding: 8px; font-size: 16px; border: 2px solid #ccc; border-radius: 5px; box-sizing: border-box;">
    </div>
    <div style="margin-bottom: 10px;">
        <label style="font-weight: bold;">
            <input name="past" value="y" type="checkbox" id="include-previous" 
                   <?php echo $show_past ? 'checked' : ''; ?> style="margin-right: 8px;">
            Incluir actividades anteriores
        </label>
    </div>
    <div>
        <button type="submit" class="button" style="margin-right: 10px;">
            <img loading="lazy" class="picto" src="/static/pictos/buscar.png"><br>Buscar
        </button>
        <a href="index.php" class="button">
            <img loading="lazy" class="picto" src="/static/pictos/cancelar.png"><br>Cancelar busqueda
        </a>
    </div>
</form>

<div id="activities-section">
<?php if (!empty($search_query)): ?>
    <h2>Resultados de búsqueda para: "<?php echo htmlspecialchars($search_query); ?>"</h2>
<?php elseif ($show_past): ?>
    <h2>Todas las actividades</h2>
<?php else: ?>
    <h2>Próximas actividades</h2>
<?php endif; ?>

<a href="crear_actividad.php" class="button" style="margin-bottom: 20px;">
    <img loading="lazy" class="picto" src="/static/pictos/mas.png"><br>
    Crear nueva actividad</a>
<br>

<div id="activities-container">
<?php if (empty($activities_to_display)): ?>
    <?php if (!empty($search_query)): ?>
        <p>No se encontraron actividades que coincidan con "<?php echo htmlspecialchars($search_query); ?>".</p>
    <?php else: ?>
        <p>No hay actividades<?php echo $show_past ? '' : ' próximas'; ?>.</p>
    <?php endif; ?>
<?php else: ?>
    <?php foreach ($activities_to_display as $activity): ?>
        <div
            style="background-color: lightcyan; padding: 10px; margin-top: 5px; border-radius: 25px; border: 2px solid black; display: inline-block;">
            <h3>
                <?php echo htmlspecialchars($activity['title']); ?>
            </h3>
            
            <?php if (isset($activity["is_shared_from"]) and $activity["is_shared_from"] != "") {?>
                <p><img loading="lazy" class="picto" src="/static/pictos/compartir2.png">
                    <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Compartido por: <br>
                        <span style="font-size: x-large;"><?php echo htmlspecialchars(iso_to_es($activity['is_shared_from'])); ?>
                        </span> </span>
                </p>
            <?php } ?>
            <p><img loading="lazy" class="picto" src="/static/pictos/dia.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Dia: <br>
                    <span style="font-size: xx-large;"><?php echo htmlspecialchars(iso_to_es($activity['start'])); ?>
                    </span> </span>
            </p>
            <p><img loading="lazy" class="picto" src="/static/pictos/inicio.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Inicio: <br>
                    <span style="font-size: xx-large;"><?php echo htmlspecialchars(explode("T", $activity['start'])[1]); ?>
                    </span> </span>
            </p>
            <p><img loading="lazy" class="picto" src="/static/pictos/fin.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Fin: <br>
                    <span style="font-size: xx-large;"><?php echo htmlspecialchars(explode("T", $activity['end'])[1]); ?>
                    </span> </span>
            </p>
            <!--<p><b>Descripción:</b> <?php echo nl2br(htmlspecialchars($activity['description'])); ?></p>-->

            <a href="actividad.php?id=<?php echo urlencode($activity['id']); ?>&global=<?php echo $activity['_global'] ? '1' : '0'; ?>"
                class="button">
                <img loading="lazy" class="picto" src="/static/pictos/leer.png"><br>Leer detalles
            </a>
            <a href="editar_actividad.php?id=<?php echo urlencode($activity['id']); ?>&global=<?php echo $activity['_global'] ? '1' : '0'; ?>"
                class="button">

                <img loading="lazy" class="picto" src="/static/pictos/escribir.png"><br>Editar
            </a>
            <a href="eliminar_actividad.php?id=<?php echo urlencode($activity['id']); ?>&global=<?php echo $activity['_global'] ? '1' : '0'; ?>"
                onclick="return confirm('¿Estás seguro?');" class="button rojo">
                <img loading="lazy" class="picto" src="/static/pictos/borrar.png"><br>Borrar</a>
        </div>
    <?php endforeach; ?>
<?php endif; ?>
</div>
</div>


<?php
require_once "../_incl/post-body.php";
?>
