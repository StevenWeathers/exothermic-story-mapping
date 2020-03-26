<script>
    export let user = {}
    export let showBorder = 'true'

    $: borderClasses = showBorder === 'true' ? 'border-b border-gray-500' : ''
</script>

<div
    class="{borderClasses} p-2 flex items-center"
    data-testId="userCard"
    data-userName="{user.name}">
    <img
        src="https://api.adorable.io/avatars/32/{user.id}.png"
        alt="Placeholder Avatar" />
    <p
        class="ml-2 text-l font-bold leading-tight truncate"
        data-testId="userName"
        title="{user.name}">
        {user.name}
    </p>
</div>
