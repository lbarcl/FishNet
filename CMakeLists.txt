cmake_minimum_required(VERSION 3.16)
project(FishNet LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_WINDOWS_EXPORT_ALL_SYMBOLS OFF)

# -------------------------------
# FishNet library
# -------------------------------
add_library(fishnet SHARED
    src/main.cpp
    src/tcpserver.cpp
)

target_include_directories(fishnet
    PUBLIC
        ${PROJECT_SOURCE_DIR}/includes
)

# -------------------------------
# Platform-specific networking
# -------------------------------
if (WIN32)
    target_link_libraries(fishnet PRIVATE ws2_32)
else()
    target_link_libraries(fishnet PRIVATE pthread)
endif()

# -------------------------------
# Test executable
# -------------------------------
add_executable(fishnet_test
    test/main.cpp
)

target_link_libraries(fishnet_test
    PRIVATE
        fishnet
)

# -------------------------------
# Output folder
# -------------------------------
set_target_properties(fishnet fishnet_test PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/bin
    LIBRARY_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/bin
    ARCHIVE_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/bin
)
