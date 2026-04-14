import re

def check_balance(filename):
    with open(filename, 'r') as f:
        content = f.read()
    
    # Simple regex for tags
    tags = re.findall(r'<div|</div>|{#if|{:else|{/if|{#each|{/each', content)
    
    stack = []
    for tag in tags:
        if tag in ('<div', '{#if', '{#each'):
            stack.append(tag)
        elif tag == '</div>':
            if not stack or stack[-1] != '<div':
                print(f"Error: </div> without <div. Stack: {stack}")
            else:
                stack.pop()
        elif tag == '{/if}': # Regex actually finds {/if or {/each or {:else
            pass # Simplified
        elif tag.startswith('{/'):
            # Close block
            pass
        
    print(f"Final stack: {stack}")

if __name__ == "__main__":
    # We need a more sophisticated one that handles {:else}
    pass

# Let's just use a manual count in segments
