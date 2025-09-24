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

// Create search index
$search_index = [];
foreach ($all_activities as $activity) {
    $search_index[] = [
        'id' => $activity['id'],
        'title' => $activity['title'],
        'description' => $activity['description'],
        'start' => $activity['start'],
        'end' => $activity['end'],
        '_global' => $activity['_global']
    ];
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

// Combine all activities for complete list
$all_activities_sorted = array_merge($upcoming_activities, $past_activities);
?>

<!-- Search functionality -->
<script>
const searchIndex = <?php echo json_encode($search_index); ?>;
const allActivities = <?php echo json_encode($all_activities_sorted); ?>;

function performSearch() {
    const searchTerm = document.getElementById('activity-search').value.toLowerCase();
    const includePrevious = document.getElementById('include-previous').checked;
    
    // Get reference to activities container
    const upcomingContainer = document.getElementById('upcoming-activities-container');
    const previousContainer = document.getElementById('previous-activities-container');
    
    // If no search term and not including previous, show default upcoming activities
    if (!searchTerm && !includePrevious) {
        location.reload();
        return;
    }
    
    // Filter activities based on search and date preference
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    let filteredActivities = allActivities.filter(activity => {
        const activityDate = new Date(activity.start);
        const matchesSearch = !searchTerm || 
            activity.title.toLowerCase().includes(searchTerm) ||
            activity.description.toLowerCase().includes(searchTerm);
        
        if (!includePrevious) {
            return matchesSearch && activityDate >= today;
        }
        return matchesSearch;
    });
    
    // Update the display
    displayFilteredActivities(filteredActivities, includePrevious);
}

function displayFilteredActivities(activities, includePrevious) {
    const upcomingContainer = document.getElementById('upcoming-activities-container');
    const previousContainer = document.getElementById('previous-activities-container');
    const upcomingSection = document.getElementById('upcoming-activities-section');
    const previousSection = document.getElementById('previous-activities-section');
    
    // Clear existing content
    upcomingContainer.innerHTML = '';
    previousContainer.innerHTML = '';
    
    if (activities.length === 0) {
        upcomingContainer.innerHTML = '<p>No se encontraron actividades que coincidan con la búsqueda.</p>';
        previousSection.style.display = 'none';
        return;
    }
    
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    let upcomingActivities = [];
    let pastActivities = [];
    
    activities.forEach(activity => {
        const activityDate = new Date(activity.start);
        if (activityDate >= today) {
            upcomingActivities.push(activity);
        } else {
            pastActivities.push(activity);
        }
    });
    
    // Display upcoming activities
    if (upcomingActivities.length > 0) {
        upcomingActivities.forEach(activity => {
            upcomingContainer.innerHTML += createActivityHTML(activity);
        });
    } else {
        upcomingContainer.innerHTML = '<p>No hay próximas actividades que coincidan con la búsqueda.</p>';
    }
    
    // Display previous activities if checkbox is checked
    if (includePrevious) {
        previousSection.style.display = 'block';
        if (pastActivities.length > 0) {
            pastActivities.forEach(activity => {
                previousContainer.innerHTML += createActivityHTML(activity, true);
            });
        } else {
            previousContainer.innerHTML = '<p>No hay actividades anteriores que coincidan con la búsqueda.</p>';
        }
    } else {
        previousSection.style.display = 'none';
    }
}

function createActivityHTML(activity, isPast = false) {
    const backgroundColor = isPast ? 'lightgray' : 'lightcyan';
    
    return `
        <div style="background-color: ${backgroundColor}; padding: 10px; margin-top: 5px; border-radius: 25px; border: 2px solid black; display: inline-block;">
            <h3>${escapeHtml(activity.title)}</h3>
            <p><img loading="lazy" class="picto" src="/static/pictos/dia.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Dia: <br>
                    <span style="font-size: xx-large;">${formatDate(activity.start)}</span> 
                </span>
            </p>
            <p><img loading="lazy" class="picto" src="/static/pictos/inicio.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Inicio: <br>
                    <span style="font-size: xx-large;">${activity.start.split('T')[1]}</span> 
                </span>
            </p>
            <p><img loading="lazy" class="picto" src="/static/pictos/fin.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Fin: <br>
                    <span style="font-size: xx-large;">${activity.end.split('T')[1]}</span> 
                </span>
            </p>
            <p><b>Descripción:</b> ${escapeHtml(activity.description).replace(/\n/g, '<br>')}</p>

            <a href="editar_actividad.php?id=${encodeURIComponent(activity.id)}&global=${activity._global ? '1' : '0'}" class="button">
                <img loading="lazy" class="picto" src="/static/pictos/escribir.png"><br>Editar
            </a>
            <a href="eliminar_actividad.php?id=${encodeURIComponent(activity.id)}&global=${activity._global ? '1' : '0'}" 
                onclick="return confirm('¿Estás seguro?');" class="button rojo">
                <img loading="lazy" class="picto" src="/static/pictos/borrar.png"><br>Borrar
            </a>
        </div>
    `;
}

function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('es-ES');
}

// Event listeners
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('activity-search');
    const checkbox = document.getElementById('include-previous');
    
    if (searchInput) {
        searchInput.addEventListener('input', performSearch);
    }
    
    if (checkbox) {
        checkbox.addEventListener('change', performSearch);
    }
});
</script>

<!-- Search Bar and Controls -->
<div style="background-color: #f9f9f9; padding: 15px; margin-bottom: 20px; border-radius: 10px; border: 1px solid #ddd;">
    <div style="margin-bottom: 10px;">
        <label for="activity-search" style="font-weight: bold; display: block; margin-bottom: 5px;">
            🔍 Buscar actividades:
        </label>
        <input type="text" id="activity-search" placeholder="Buscar por título o descripción..." 
               style="width: 100%; padding: 8px; font-size: 16px; border: 2px solid #ccc; border-radius: 5px; box-sizing: border-box;">
    </div>
    <div>
        <label style="font-weight: bold;">
            <input type="checkbox" id="include-previous" style="margin-right: 8px;">
            Incluir actividades anteriores
        </label>
    </div>
</div>

<div id="upcoming-activities-section">
<h2>Próximas Actividades</h2>

<a href="crear_actividad.php" class="button" style="margin-bottom: 20px;">
    <img loading="lazy" class="picto" src="/static/pictos/mas.png"><br>
    Crear nueva actividad</a>
<br>

<div id="upcoming-activities-container">
<?php if (empty($upcoming_activities)): ?>
    <p>No hay próximas actividades.</p>
<?php else: ?>
    <?php foreach ($upcoming_activities as $activity): ?>
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
            <p><b>Descripción:</b> <?php echo nl2br(htmlspecialchars($activity['description'])); ?></p>

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

<div id="previous-activities-section" style="display: none;">
<h2>Actividades Anteriores</h2>
<div id="previous-activities-container">
    <!-- Previous activities will be populated by JavaScript -->
</div>
</div>

<?php
require_once "../_incl/post-body.php";
?>
