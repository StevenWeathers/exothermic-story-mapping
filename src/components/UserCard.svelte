<script>
    export let user = {}
</script>

<div
    class="border-b border-gray-500 p-4 flex items-center"
    data-testId="userCard"
    data-userName="{user.name}">
    <div class="w-1/4">
        <img
            src="https://api.adorable.io/avatars/48/{user.id}.png"
            alt="Placeholder Avatar" />
    </div>
    <div class="w-3/4">
        <div class="flex items-center">
            <p
                class="text-xl font-bold leading-tight truncate"
                data-testId="userName"
                title="{user.name}">
                {user.name}
            </p>
        </div>
    </div>
</div>
