const socket = io("http://localhost:8000", {
    transports: ["websocket"],
});
const params = new URLSearchParams(window.location.search);
let token = params.get("token");

async function enter() {
    const status = localStorage.getItem("status");
    if (status === "queued") {
        return true;
    }

    if (status === "active") {
        window.location.href = "https://google.com";
        return false;
    }
    
    const res = await fetch("http://localhost:8000/api/v1/queues/enter", {
        method: "POST",
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        }
    });
    if (res.status === 307) {
        localStorage.setItem("status", "active");
        window.location.href = "https://google.com";

        return false;
    }

    if (res.status === 401) {
        await refreshToken();

        return enter();
    }

    localStorage.setItem("status", "queued");

    return true;
}

async function getPosition() {
    const res = await fetch("http://localhost:8000/api/v1/queues/position", {
        method: "GET",
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        }
    });
    if (res.status === 401) {
        await refreshToken();
        return;
    }

    const data = await res.json();
    document.getElementById("position").innerText = data.position;
}

async function getAvg() {
    const res = await fetch("http://localhost:8000/api/v1/queues/status", {
        method: "GET",
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        }
    });
    if (res.status === 401) {
        await refreshToken();
        return;
    }

    const data = await res.json();
    document.getElementById("avg").innerText = formatTime(data.average_queue_time);
}

function formatTime(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secondsRemaining = seconds % 60;

    let parts = [];
    if (hours > 0) parts.push(`${hours} hora${hours > 1 ? 's' : ''}`);
    if (minutes > 0) parts.push(`${minutes} minuto${minutes > 1 ? 's' : ''}`);
    if (secondsRemaining > 0 || parts.length === 0) {
        parts.push(`${secondsRemaining} segundo${secondsRemaining > 1 ? 's' : ''}`);
    }

    return parts.join(", ");
}

async function refreshToken() {
    const res = await fetch("http://localhost:8000/api/v1/queues/refresh-token", {
        method: "POST",
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        }
    });
    const data = await res.json();
    token = data.token;
}

socket.on("connect", () => {
    if (token) {
        socket.emit("join", token);
    }
});

socket.on("releaseEntry", () => {
    localStorage.setItem("status", "active");
    window.location.href = "https://google.com";
});

isEntered = enter();
if (isEntered) {
    setInterval(getPosition, 2000);
    setInterval(getAvg, 5000);
}
