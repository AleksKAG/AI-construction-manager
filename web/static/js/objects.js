// web/static/js/objects.js

import { apiFetch } from './auth.js';

let objects = [];

export async function loadObjects() {
    objects = await apiFetch('/objects');
    renderObjects();
}

export function renderObjects() {
    const tbody = document.getElementById('objectsBody');
    if (!tbody) return;
    
    tbody.innerHTML = objects.map(obj => `
        <tr>
            <td><a href="/object/${obj.id}">${obj.name}</a></td>
            <td>${obj.address || '—'}</td>
            <td>${obj.type || '—'}</td>
            <td><span class="badge badge-${getStatusClass(obj.status)}">${obj.status || '—'}</span></td>
            <td class="flex-center">
                <button class="btn-secondary" onclick="editObject('${obj.id}')">✏️</button>
                <button class="btn-danger" onclick="deleteObject('${obj.id}')">🗑️</button>
            </td>
        </tr>
    `).join('');
}

function getStatusClass(status) {
    const map = {
        'проектирование': 'warning',
        'строительство': 'success',
        'сдан': 'success',
        'приостановлен': 'danger',
    };
    return map[status] || 'secondary';
}

export async function createObject(data) {
    const response = await apiFetch('/objects', {
        method: 'POST',
        body: JSON.stringify(data),
    });
    objects.push(response);
    renderObjects();
    return response;
}

export async function updateObject(id, data) {
    const response = await apiFetch(`/objects/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
    });
    const index = objects.findIndex(o => o.id === id);
    if (index !== -1) objects[index] = response;
    renderObjects();
    return response;
}

export async function deleteObject(id) {
    if (!confirm('Удалить объект?')) return;
    
    await apiFetch(`/objects/${id}`, { method: 'DELETE' });
    objects = objects.filter(o => o.id !== id);
    renderObjects();
}

// Глобальные функции для inline onclick
window.editObject = function(id) {
    const obj = objects.find(o => o.id === id);
    if (!obj) return;
    
    document.getElementById('modalTitle').textContent = 'Редактировать объект';
    document.getElementById('objectId').value = obj.id;
    document.getElementById('objName').value = obj.name;
    document.getElementById('objAddress').value = obj.address || '';
    document.getElementById('objType').value = obj.type || 'новое строительство';
    document.getElementById('objStatus').value = obj.status || 'проектирование';
    document.getElementById('modal').classList.remove('hidden');
};

window.deleteObject = deleteObject;
