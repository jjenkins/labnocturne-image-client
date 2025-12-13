"""Setup configuration for labnocturne package"""

from setuptools import setup, find_packages
from pathlib import Path

# Read the README file
this_directory = Path(__file__).parent
long_description = (this_directory / "README.md").read_text() if (this_directory / "README.md").exists() else ""

setup(
    name="labnocturne",
    version="1.0.0",
    author="Lab Nocturne",
    author_email="support@labnocturne.com",
    description="Python client for Lab Nocturne Images API",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/jjenkins/labnocturne-image-client",
    project_urls={
        "Bug Tracker": "https://github.com/jjenkins/labnocturne-image-client/issues",
        "Documentation": "https://github.com/jjenkins/labnocturne-image-client/tree/main/python",
        "Source Code": "https://github.com/jjenkins/labnocturne-image-client/tree/main/python",
    },
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
    python_requires=">=3.7",
    install_requires=[
        "requests>=2.25.0",
    ],
    extras_require={
        "async": ["aiohttp>=3.8.0"],
        "dev": [
            "pytest>=7.0.0",
            "black>=22.0.0",
            "flake8>=4.0.0",
            "mypy>=0.950",
        ],
    },
    keywords="images cdn upload storage api labnocturne",
    license="MIT",
)
