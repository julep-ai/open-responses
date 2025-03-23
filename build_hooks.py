#!/usr/bin/env python3
import os
import platform
import shutil
from hatchling.builders.hooks.plugin.interface import BuildHookInterface

# Paths to binaries relative to project root
GO_BINARIES = {
    'linux': 'bin/open-responses-linux',
    'darwin': 'bin/open-responses-macos',
    'win32': 'bin/open-responses-win.exe'
}

class CustomBuildHook(BuildHookInterface):
    def initialize(self, version, build_data):
        """Hook called before build to prepare files"""
        build_dir = self.root
        
        # Create the bin directory inside the package
        os.makedirs(os.path.join(build_dir, 'open_responses', 'bin'), exist_ok=True)
        
        # Copy all binaries to the package
        for platform_name, binary_path in GO_BINARIES.items():
            src_path = os.path.join(build_dir, binary_path)
            if os.path.exists(src_path):
                dst_path = os.path.join(build_dir, 'open_responses', 'bin', os.path.basename(binary_path))
                print(f"Copying {src_path} to {dst_path}")
                shutil.copy(src_path, dst_path)
                # Make executable
                if os.name != 'nt':  # Not Windows
                    os.chmod(dst_path, 0o755)
        
        return True
