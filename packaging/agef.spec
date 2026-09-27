Name:           agef
Version:        1.0.0
Release:        1%{?dist}
Summary:        High-performance recursive folder encryption and decryption using age
License:        MIT
URL:            https://github.com/user/agef
BuildArch:      x86_64

Source0:        agef

%description
agef is a fast, standalone folder encryption and decryption tool
powered by the age encryption library. It allows mass encryption and decryption
of entire folder trees while preserving directory hierarchy and timestamps,
creating in-place encrypted and decrypted folders.

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{_bindir}
install -m 0755 %{SOURCE0} %{buildroot}%{_bindir}/agef

%files
%{_bindir}/agef

%changelog
* Sun Sep 27 2026 Open Source Contributor <contributor@example.com> - 1.0.0-1
- Initial release of agef with native Go age integration
