import os
import sys

def output_directory_contents(directories):
    output_file = "output.txt"

    with open(output_file, 'w', encoding='utf-8') as out:
        for base_dir in directories:
            base_dir = base_dir.rstrip('/')

            for root, _, files in os.walk(base_dir):
                relative_path = os.path.relpath(root, base_dir)
                if relative_path == ".":
                    relative_path = ""

                for file in files:
                    file_path = os.path.join(root, file)
                    try:
                        with open(file_path, 'r', encoding='utf-8') as f:
                            content = f.read()
                        out.write(f"--- {os.path.join(base_dir, relative_path, file)} ---\n")
                        out.write(content + "\n\n")
                    except Exception as e:
                        out.write(f"--- {os.path.join(base_dir, relative_path, file)} (読み込み失敗: {e}) ---\n\n")

if __name__ == "__main__":
    directories = [line.strip() for line in sys.stdin if line.strip()]
    output_directory_contents(directories)
