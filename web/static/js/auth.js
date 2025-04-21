// Form configuration defaults
const formDefaults = {
    responseType: 'token',
    redirectUri: 'http://localhost:3000/twitch/callback',
    defaultScopes: [
        'chat:read',
        'chat:edit',
    ],
    availableScopes: [
        'channel:manage:polls',
        'channel:read:polls',
        'channel:read:redemptions',
        'channel:manage:redemptions',
        'moderator:read:chatters',
        'moderator:manage:chat_messages',
        'chat:read',
        'chat:edit',
        'channel:moderate',
        'channel:read:subscriptions'
    ]
};

function updateScopeValue() {
    const checkboxes = document.querySelectorAll('input[name="scope_checkbox"]:checked');
    const hiddenInput = document.getElementById('scope_value');
    const selectedValues = Array.from(checkboxes).map(checkbox => checkbox.value);
    hiddenInput.value = selectedValues.join(' ');
}

// Create scope checkbox element
function createScopeCheckbox(scope, isChecked) {
    const div = document.createElement('div');
    div.className = 'flex items-center';
    
    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.name = 'scope_checkbox';
    checkbox.value = scope;
    checkbox.className = 'h-4 w-4 text-[#6441a5] focus:ring-[#6441a5] border-gray-300 rounded';
    checkbox.checked = isChecked;
    checkbox.onchange = updateScopeValue;
    
    const label = document.createElement('label');
    label.className = 'ml-2 block text-sm text-gray-700';
    label.textContent = scope;
    
    div.appendChild(checkbox);
    div.appendChild(label);
    return div;
}

// Initialize form with defaults
function initializeForm() {
    // Set response type
    document.querySelector('input[name="response_type"]').value = formDefaults.responseType;
    
    // Set redirect URI
    document.querySelector('input[name="redirect_uri"]').value = formDefaults.redirectUri;
    
    // Initialize scope checkboxes
    const scopeContainer = document.getElementById('scope_container');
    formDefaults.availableScopes.forEach(scope => {
        const isChecked = formDefaults.defaultScopes.includes(scope);
        const checkboxElement = createScopeCheckbox(scope, isChecked);
        scopeContainer.appendChild(checkboxElement);
    });
    
    // Initialize scope value
    updateScopeValue();
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', initializeForm); 